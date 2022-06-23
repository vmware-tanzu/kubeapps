// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Inspired by
// https://github.com/fluxcd/source-controller/blob/main/internal/helm/repository/oci_chart_repository.go
// and adapted for kubeapps use.
// OCI spec ref
//   https://github.com/opencontainers/distribution-spec/blob/main/spec.md

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/semver/v3"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common/transport"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	log "k8s.io/klog/v2"

	"github.com/fluxcd/pkg/version"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"

	// OCI Registry As a Storage (ORAS)
	registryauth "oras.land/oras-go/pkg/registry/remote/auth"
)

// RegistryClient is an interface for interacting with OCI registries
// It is used by the OCIRegistry to retrieve chart versions
// from OCI registries. These signatures are implemented by
// https://github.com/helm/helm/blob/main/pkg/registry/client.go
type RegistryClient interface {
	Login(host string, opts ...registry.LoginOption) error
	Logout(host string, opts ...registry.LogoutOption) error
	Tags(url string) ([]string, error)
}

// an interface flux plugin uses to determine what kind of vendor-specific
// registry repository name lister applies, and then executes specific logic
type OCIRepositoryLister interface {
	IsApplicableFor(*OCIRegistry) (bool, error)
	ListRepositoryNames() ([]string, error)
}

// OCIRegistry represents a Helm chart repository, and the configuration
// required to download the repository tags and charts from the repository.
// All methods are thread safe unless defined otherwise.
type OCIRegistry struct {
	// url is the location of the repository.
	url url.URL
	// helmGetter to use while accessing the repository's contents.
	helmGetter getter.Getter
	// helmOptions to configure the Client with while downloading tags
	// or a chart from the URL.
	helmOptions []getter.Option

	tlsConfig *tls.Config

	// registryClient is a client to use while downloading tags or charts from a registry.
	registryClient RegistryClient

	// The set of public operations one can use w.r.t. RegistryClient is very small
	// (Login/Logout/Tags). I need to be able to query remote OCI repo for ListRepositoryNames(),
	// which is not in the set and all fields of RegistryClient,
	//  including repositoryAuthorizer are internal, so this is a workaround
	registryCredentialFn OCIRegistryCredentialFn

	repositoryLister OCIRepositoryLister
}

// OCIRegistryOption is a function that can be passed to NewOCIRegistry
// to configure an OCIRegistry.
type OCIRegistryOption func(*OCIRegistry) error

type OCIRegistryCredentialFn func(ctx context.Context, reg string) (registryauth.Credential, error)

var (
	helmGetters = getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
		getter.Provider{
			Schemes: []string{"oci"},
			New:     getter.NewOCIGetter,
		},
	}

	// TODO: make this thing extensible so code coming from other plugs/modules
	// can register new repository listers
	builtInRepoListers = []OCIRepositoryLister{
		NewGitHubRepositoryLister(),
		// TODO
	}
)

// withOCIRegistryClient returns a OCIRegistryOption that will set the registry client
func withOCIRegistryClient(client RegistryClient) OCIRegistryOption {
	return func(r *OCIRegistry) error {
		r.registryClient = client
		return nil
	}
}

// withOCIGetter returns a OCIRegistryOption that will set the getter.Getter
func withOCIGetter(providers getter.Providers) OCIRegistryOption {
	return func(r *OCIRegistry) error {
		c, err := providers.ByScheme(r.url.Scheme)
		if err != nil {
			return err
		}
		r.helmGetter = c
		return nil
	}
}

// withOCIGetterOptions returns a OCIRegistryOption that will set the getter.Options
func withOCIGetterOptions(getterOpts []getter.Option) OCIRegistryOption {
	return func(r *OCIRegistry) error {
		r.helmOptions = getterOpts
		return nil
	}
}

func withRegistryCredentialFn(fn OCIRegistryCredentialFn) OCIRegistryOption {
	return func(r *OCIRegistry) error {
		r.registryCredentialFn = fn
		return nil
	}
}

// newOCIRegistry constructs and returns a new OCIRegistry with
// the RegistryClient configured to the getter.Getter for the
// registry URL scheme. It returns an error on URL parsing failures.
// It assumes that the url scheme has been validated to be an OCI scheme.
func newOCIRegistry(registryURL string, registryOpts ...OCIRegistryOption) (*OCIRegistry, error) {
	u, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}

	r := &OCIRegistry{}
	r.url = *u
	for _, opt := range registryOpts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *OCIRegistry) listRepositoryNames() ([]string, error) {
	// this needs to be done after a call to login()
	for _, lister := range builtInRepoListers {
		if ok, err := lister.IsApplicableFor(r); ok && err == nil {
			r.repositoryLister = lister
			break
		} else {
			log.Infof("Lister [%v] not applicable for registry for URL: [%s] [%v]", reflect.TypeOf(lister), &r.url, err)
		}
	}

	if r.repositoryLister == nil {
		return nil, status.Errorf(codes.Internal, "No repository lister found for OCI registry with url: [%s]", &r.url)
	}

	return r.repositoryLister.ListRepositoryNames()
}

// Get returns the ChartVersion for the given name, the version is expected
// to be a semver.Constraints compatible string. If version is empty, the latest
// stable version will be returned and prerelease versions will be ignored.
// adapted from https://github.com/helm/helm/blob/49819b4ef782e80b0c7f78c30bd76b51ebb56dc8/pkg/downloader/chart_downloader.go#L162
func (r *OCIRegistry) getChartVersion(name, ver string) (*repo.ChartVersion, error) {
	log.Infof("+getChartVersion(%s,%s", name, ver)
	// Find chart versions matching the given name.
	// Either in an index file or from a registry.
	ref := fmt.Sprintf("%s/%s", r.url.String(), name)
	cvs, err := r.getTags(ref)
	if err != nil {
		return nil, err
	}

	if len(cvs) == 0 {
		return nil, status.Errorf(codes.Internal, "unable to locate any tags in provided repository: %s", name)
	}

	// Determine if version provided
	// If empty, try to get the highest available tag
	// If exact version, try to find it
	// If semver constraint string, try to find a match
	tag, err := getLastMatchingVersionOrConstraint(cvs, ver)
	return &repo.ChartVersion{
		URLs: []string{fmt.Sprintf("%s/%s:%s", r.url.String(), name, tag)},
		Metadata: &chart.Metadata{
			Name:    name,
			Version: tag,
		},
	}, err
}

// This function shall be called for OCI registries only
// It assumes that the ref has been validated to be an OCI reference.
func (r *OCIRegistry) getTags(ref string) ([]string, error) {
	ref = strings.TrimPrefix(ref, fmt.Sprintf("%s://", registry.OCIScheme))

	tags, err := r.registryClient.Tags(ref)
	if err != nil {
		return nil, err
	}

	log.Infof("getTags(%s): %s %v", ref, tags, err)
	if len(tags) == 0 {
		return nil, fmt.Errorf("unable to locate any tags in provided repository: %s", ref)
	}
	return tags, nil
}

// downloadChart confirms the given repo.ChartVersion has a downloadable URL,
// and then attempts to download the chart using the Client and Options of the
// OCIRegistry. It returns a bytes.Buffer containing the chart data.
// In case of an OCI hosted chart, this function assumes that the chartVersion url is valid.
func (r *OCIRegistry) downloadChart(chart *repo.ChartVersion) (*bytes.Buffer, error) {
	if len(chart.URLs) == 0 {
		return nil, fmt.Errorf("chart '%s' has no downloadable URLs", chart.Name)
	}

	ref := chart.URLs[0]
	u, err := url.Parse(ref)
	if err != nil {
		err = fmt.Errorf("invalid chart URL format '%s': %w", ref, err)
		return nil, err
	}

	t := transport.NewOrIdle(r.tlsConfig)
	clientOpts := append(r.helmOptions, getter.WithTransport(t))
	defer transport.Release(t)

	// trim the oci scheme prefix if needed
	getThis := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	return r.helmGetter.Get(getThis, clientOpts...)
}

// logout attempts to logout from the OCI registry.
// It returns an error on failure.
func (r *OCIRegistry) logout() error {
	log.Infof("+logout")
	err := r.registryClient.Logout(r.url.Host)
	if err != nil {
		return err
	}
	return nil
}

// getLastMatchingVersionOrConstraint returns the last version that matches the given version string.
// If the version string is empty, the highest available version is returned.
func getLastMatchingVersionOrConstraint(cvs []string, ver string) (string, error) {
	// Check for exact matches first
	if ver != "" {
		for _, cv := range cvs {
			if ver == cv {
				return cv, nil
			}
		}
	}

	// Continue to look for a (semantic) version match
	verConstraint, err := semver.NewConstraint("*")
	if err != nil {
		return "", err
	}
	latestStable := ver == "" || ver == "*"
	if !latestStable {
		verConstraint, err = semver.NewConstraint(ver)
		if err != nil {
			return "", err
		}
	}

	matchingVersions := make([]*semver.Version, 0, len(cvs))
	for _, cv := range cvs {
		v, err := version.ParseVersion(cv)
		if err != nil {
			continue
		}

		if !verConstraint.Check(v) {
			continue
		}

		matchingVersions = append(matchingVersions, v)
	}
	if len(matchingVersions) == 0 {
		return "", fmt.Errorf("could not locate a version matching provided version string %s", ver)
	}

	// Sort versions
	sort.Sort(sort.Reverse(semver.Collection(matchingVersions)))

	return matchingVersions[0].Original(), nil
}

// newRegistryClient generates a registry client and a temporary credential file.
// The client is meant to be used for a single reconciliation.
// The file is meant to be used for a single reconciliation and deleted after.
func newRegistryClient(isLogin bool) (*registry.Client, string, error) {
	if isLogin {
		// create a temporary file to store the credentials
		// this is needed because otherwise the credentials are stored in ~/.docker/config.json.
		credentialFile, err := os.CreateTemp("", "credentials")
		if err != nil {
			return nil, "", err
		}

		clientOpts := []registry.ClientOption{
			registry.ClientOptWriter(io.Discard),
			registry.ClientOptCredentialsFile(credentialFile.Name()),
		}
		rClient, err := registry.NewClient(clientOpts...)
		if err != nil {
			return nil, "", err
		}
		return rClient, credentialFile.Name(), nil
	}

	clientOpts := []registry.ClientOption{
		registry.ClientOptWriter(io.Discard),
		registry.ClientOptDebug(true),
	}

	rClient, err := registry.NewClient(clientOpts...)
	if err != nil {
		return nil, "", err
	}
	return rClient, "", nil
}

// OCI Helm repository, which defines a source, does not produce an Artifact
// ref https://fluxcd.io/docs/components/source/helmrepositories/#helm-oci-repository
func (s *repoEventSink) onAddOciRepo(repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onAddOciRepo(%s)", common.PrettyPrint(repo))
	defer log.Infof("-onAddOciRepo")

	ociRegistry, err := s.newOCIRegistryAndLogin(repo)
	if err != nil {
		return nil, false, err
	}

	chartRepo := &models.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Spec.URL,
		Type:      repo.Spec.Type,
	}

	// repository names aka application names
	appNames, err := ociRegistry.listRepositoryNames()
	if err != nil {
		return nil, false, err
	}

	charts := []models.Chart{}
	for _, appName := range appNames {
		log.Infof("==========>: app name: [%s]", appName)

		chartVersion, err := ociRegistry.getChartVersion(appName, "")
		if err != nil {
			return nil, false, status.Errorf(codes.Internal, "%v", err)
		}
		log.Infof("==========>: chart version: %s", common.PrettyPrint(chartVersion))

		chartBuffer, err := ociRegistry.downloadChart(chartVersion)
		if err != nil {
			return nil, false, status.Errorf(codes.Internal, "%v", err)
		}
		log.Infof("==========>: chart buffer: [%d] bytes", chartBuffer.Len())

		// Encode repository names to store them in the database.
		encodedAppName := url.PathEscape(appName)

		// not sure yet why flux untars into a temp directory
		chartID := path.Join(repo.Name, encodedAppName)
		files, err := tarutil.FetchChartDetailFromTarball(
			bytes.NewReader(chartBuffer.Bytes()), chartID)
		if err != nil {
			return nil, false, status.Errorf(codes.Internal, "%v", err)
		}

		log.Infof("==========>: files: [%d]", len(files))

		chartYaml, ok := files[models.ChartYamlKey]
		// TODO (gfichtenholt): if there is no chart yaml (is that even possible?),
		// fall back to chart info from repo index.yaml
		if !ok || chartYaml == "" {
			return nil, false, status.Errorf(codes.Internal, "No chart manifest found for chart [%s]", chartID)
		}
		var chartMetadata chart.Metadata
		err = yaml.Unmarshal([]byte(chartYaml), &chartMetadata)
		if err != nil {
			return nil, false, err
		}

		log.Infof("==========>: chart metadata: %s", common.PrettyPrint(chartMetadata))

		maintainers := []chart.Maintainer{}
		for _, maintainer := range chartMetadata.Maintainers {
			maintainers = append(maintainers, *maintainer)
		}

		modelsChartVersion := models.ChartVersion{
			Version:    chartVersion.Version,
			AppVersion: chartVersion.AppVersion,
			Created:    chartVersion.Created,
			Digest:     chartVersion.Digest,
			URLs:       chartVersion.URLs,
			Readme:     files[models.ReadmeKey],
			Values:     files[models.ValuesKey],
			Schema:     files[models.SchemaKey],
		}

		chart := models.Chart{
			ID:          path.Join(repo.Name, encodedAppName),
			Name:        encodedAppName,
			Repo:        chartRepo,
			Description: chartMetadata.Description,
			Home:        chartMetadata.Home,
			Keywords:    chartMetadata.Keywords,
			Maintainers: maintainers,
			Sources:     chartMetadata.Sources,
			Icon:        chartMetadata.Icon,
			Category:    chartMetadata.Annotations["category"],
			ChartVersions: []models.ChartVersion{
				modelsChartVersion,
			},
		}
		charts = append(charts, chart)
	}

	checksum, err := ociRegistry.checksum()
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "%v", err)
	}

	cacheEntryValue := repoCacheEntryValue{
		Checksum: checksum,
		Charts:   charts,
	}

	// use gob encoding instead of json, it peforms much better
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(cacheEntryValue); err != nil {
		return nil, false, err
	}
	return buf.Bytes(), true, nil
}

func (s *repoEventSink) onModifyOciRepo(key string, oldValue interface{}, repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onModifyOciRepo(%s)", common.PrettyPrint(repo))
	defer log.Infof("-onModifyOciRepo")

	// We should to compare checksums on what's stored in the cache
	// vs the modified object to see if the contents has really changed before embarking on
	// an expensive operation
	cacheEntryUntyped, err := s.onGetRepo(key, oldValue)
	if err != nil {
		return nil, false, err
	}

	cacheEntry, ok := cacheEntryUntyped.(repoCacheEntryValue)
	if !ok {
		return nil, false, status.Errorf(
			codes.Internal,
			"unexpected value found in cache for key [%s]: %v",
			key, cacheEntryUntyped)
	}

	ociRegistry, err := s.newOCIRegistryAndLogin(repo)
	if err != nil {
		return nil, false, err
	}

	newChecksum, err := ociRegistry.checksum()
	if err != nil {
		return nil, false, err
	}

	if cacheEntry.Checksum != newChecksum {
		return nil, false, status.Errorf(codes.Unimplemented, "TODO")
	} else {
		// skip because the content did not change
		return nil, false, nil
	}
}

//
// misc OCI repo utilities
//

// TagList represents a list of tags as specified at
// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#content-discovery
type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// Checksum returns the sha256 of the repo by concatenating tags for
// all repositories within the registry and returning the sha256.
// Caveat: Mutated image tags won't be detected as new
func (r *OCIRegistry) checksum() (string, error) {
	appNames, err := r.listRepositoryNames()
	if err != nil {
		return "", err
	}
	tags := map[string]TagList{}
	for _, appName := range appNames {
		ref := fmt.Sprintf("%s/%s", r.url.String(), appName)
		tagss, err := r.getTags(ref)
		if err != nil {
			return "", err
		}
		tags[appName] = TagList{Name: appName, Tags: tagss}
	}

	content, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return common.GetSha256(content)
}

func (s *repoEventSink) newOCIRegistryAndLogin(repo sourcev1.HelmRepository) (*OCIRegistry, error) {
	// Construct the Getter options from the HelmRepository data
	loginOpts, getterOpts, cred, err := s.clientOptionsForRepo(repo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client options: %v", err)
	}
	log.Infof("=====================> loginOpts: [%v], getterOpts: [%v], cred: %t", len(loginOpts), len(getterOpts), cred != nil)

	// Create registry client and login if needed.
	registryClient, file, err := newRegistryClient(loginOpts != nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client: %v", err)
	}
	if file != "" {
		/*
			defer func() {
				byteArray, err := ioutil.ReadFile(file)
				if err != nil {
					log.Infof("Failed to read temporary credentials file [%s]: %v", file, err)
				}

				// Convert []byte to string and print to screen
				log.Infof("about to remove temporary credentials file [%s] content\n[%s]",
					file, string(byteArray))

				if err := os.Remove(file); err != nil {
					log.Infof("Failed to delete temporary credentials file: %v", err)
				}
			}()
		*/
	}

	registryCredentialFn := func(ctx context.Context, reg string) (registryauth.Credential, error) {
		if cred != nil {
			return *cred, nil
		} else {
			return registryauth.EmptyCredential, nil
		}
	}
	// a little bit misleading, since repo.Spec.URL is really an OCI Registry URL,
	// which may contain zero or more "helm repositories", such as
	// oci://demo.goharbor.io/test-oci-1, which may contain repositories "repo-1", "repo2", etc

	ociRegistry, err := newOCIRegistry(
		repo.Spec.URL,
		withOCIGetter(helmGetters),
		withOCIGetterOptions(getterOpts),
		withOCIRegistryClient(registryClient),
		withRegistryCredentialFn(registryCredentialFn))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse URL '%s': %v", repo.Spec.URL, err)
	}

	// Attempt to login to the registry if credentials are provided.
	if loginOpts != nil {
		err := ociRegistry.registryClient.Login(ociRegistry.url.Host, loginOpts...)
		log.Infof("login(%s): %v", ociRegistry.url.Host, err)
		if err != nil {
			return nil, err
		}
	}
	return ociRegistry, nil
}

func (s *repoEventSink) clientOptionsForRepo(repo sourcev1.HelmRepository) ([]registry.LoginOption, []getter.Option, *registryauth.Credential, error) {
	log.Infof("+clientOptionsForRepo()")

	var loginOpts []registry.LoginOption
	var cred *registryauth.Credential
	getterOpts := []getter.Option{
		getter.WithURL(repo.Spec.URL),
		getter.WithTimeout(repo.Spec.Timeout.Duration),
		getter.WithPassCredentialsAll(repo.Spec.PassCredentials),
	}

	secret, err := s.getRepoSecret(context.Background(), repo)
	if err != nil {
		return nil, nil, nil, err
	} else if secret != nil {
		opts, err := common.HelmGetterOptionsFromSecret(*secret)
		if err != nil {
			return nil, nil, nil, err
		}
		getterOpts = append(getterOpts, opts...)

		cred, err = common.OCIRegistryCredentialFromSecret(repo.Spec.URL, *secret)
		if err != nil {
			return nil, nil, nil, err
		}

		loginOpt := registry.LoginOptBasicAuth(cred.Username, cred.Password)
		if loginOpt != nil {
			loginOpts = append(loginOpts, loginOpt)
		}
	}
	return loginOpts, getterOpts, cred, nil
}
