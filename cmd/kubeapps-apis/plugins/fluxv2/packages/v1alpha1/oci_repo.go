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
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"

	"github.com/fluxcd/pkg/version"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"

	// OCI Registry As a Storage (ORAS)
	orasregistryauthv2 "oras.land/oras-go/v2/registry/remote/auth"
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
	ListRepositoryNames(ociRegistry *OCIRegistry) ([]string, error)
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

type OCIRegistryCredentialFn func(ctx context.Context, reg string) (orasregistryauthv2.Credential, error)

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
		NewDockerRegistryApiV2RepositoryLister(),
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
			log.Infof("Lister [%v] not applicable for registry for URL: [%s] [%v]", reflect.TypeOf(lister), r.url.String(), err)
		}
	}

	if r.repositoryLister == nil {
		return nil, status.Errorf(codes.Internal, "No repository lister found for OCI registry with url: [%s]", &r.url)
	}

	return r.repositoryLister.ListRepositoryNames(r)
}

// pickChartVersionFrom returns the ChartVersion for the given name, the version is expected
// to be a semver.Constraints compatible string. If version is empty, the latest
// stable version will be returned and prerelease versions will be ignored.
// adapted from https://github.com/helm/helm/blob/49819b4ef782e80b0c7f78c30bd76b51ebb56dc8/pkg/downloader/chart_downloader.go#L162
func (r *OCIRegistry) pickChartVersionFrom(name, ver string, cvs []string) (*repo.ChartVersion, error) {
	log.Infof("+pickChartVersionFrom(%s,%s,%s)", name, ver, cvs)

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
	log.Infof("+getTags(%s)", ref)
	defer log.Infof("-getTags(%s)", ref)

	ref = strings.TrimPrefix(ref, fmt.Sprintf("%s://", registry.OCIScheme))

	log.Infof("getTags: about to call .Tags(%s)", ref)
	tags, err := r.registryClient.Tags(ref)
	log.Infof("getTags: done with call .Tags(%s): %s %v", ref, tags, err)
	if err != nil {
		return nil, err
	}

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

	defer func() {
		err = transport.Release(t)
	}()

	if err != nil {
		return nil, err
	}

	// trim the oci scheme prefix if needed
	getThis := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	log.Infof("about to call helmGetter.Get(%s)", getThis)
	buf, err := r.helmGetter.Get(getThis, clientOpts...)
	if buf != nil {
		log.Infof("helmGetter.Get(%s) returned buffer size: [%d] error: %v", getThis, len(buf.Bytes()), err)
	} else {
		log.Infof("helmGetter.Get(%s) returned error: %v", getThis, err)
	}
	return buf, err
}

// logout attempts to logout from the OCI registry.
// It returns an error on failure.
//nolint:unused
func (r *OCIRegistry) logout() error {
	log.Info("+logout")
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

// TODO: this function is way too long. Break it up
func (s *repoEventSink) onAddOciRepo(repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onAddOciRepo(%s)", common.PrettyPrint(repo))
	defer log.Info("-onAddOciRepo")

	ociRegistry, err := s.newOCIRegistryAndLoginWithRepo(context.Background(), repo)
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
	for _, fullAppName := range appNames {
		appName, err := ociRegistry.shortRepoName(fullAppName)
		if err != nil {
			return nil, false, err
		}

		// Encode repository names to store them in the database.
		encodedAppName := url.PathEscape(appName)
		chartID := path.Join(repo.Name, encodedAppName)

		log.Infof("==========>: app name: [%s], chartID: [%s]", appName, chartID)

		ref := fmt.Sprintf("%s/%s", ociRegistry.url.String(), appName)
		allTags, err := ociRegistry.getTags(ref)
		if err != nil {
			return nil, false, err
		}

		// to be consistent with how we support helm http repos
		// the chart fields like Desciption, home, sources come from the
		// most recent chart version
		// ref https://github.com/vmware-tanzu/kubeapps/blob/11c87926d6cd798af72875d01437d15ae8d85b9a/pkg/helm/index.go#L30
		log.Infof("==========>: most recent chart version: %s", allTags[0])
		latestChartVersion, err := ociRegistry.pickChartVersionFrom(appName, allTags[0], allTags)
		if err != nil {
			return nil, false, status.Errorf(codes.Internal, "%v", err)
		}

		latestChartMetadata, err := getOCIChartMetadata(ociRegistry, chartID, latestChartVersion)
		if err != nil {
			return nil, false, err
		}

		maintainers := []chart.Maintainer{}
		for _, maintainer := range latestChartMetadata.Maintainers {
			maintainers = append(maintainers, *maintainer)
		}

		mc := models.Chart{
			ID:            chartID,
			Name:          encodedAppName,
			Repo:          chartRepo,
			Description:   latestChartMetadata.Description,
			Home:          latestChartMetadata.Home,
			Keywords:      latestChartMetadata.Keywords,
			Maintainers:   maintainers,
			Sources:       latestChartMetadata.Sources,
			Icon:          latestChartMetadata.Icon,
			Category:      latestChartMetadata.Annotations["category"],
			ChartVersions: []models.ChartVersion{},
		}

		for _, tag := range allTags {
			chartVersion, err := ociRegistry.pickChartVersionFrom(appName, tag, allTags)
			if err != nil {
				return nil, false, status.Errorf(codes.Internal, "%v", err)
			}
			log.Infof("==========>: chart version: %s", common.PrettyPrint(chartVersion))

			mcv := models.ChartVersion{
				Version:    chartVersion.Version,
				AppVersion: chartVersion.AppVersion,
				Created:    chartVersion.Created,
				Digest:     chartVersion.Digest,
				URLs:       chartVersion.URLs,
			}
			mc.ChartVersions = append(mc.ChartVersions, mcv)
		}
		charts = append(charts, mc)
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

	if s.chartCache != nil {
		fn := downloadOCIChartFn(ociRegistry)
		if err = s.chartCache.SyncCharts(charts, fn); err != nil {
			return nil, false, err
		}
	}

	return buf.Bytes(), true, nil
}

func (s *repoEventSink) onModifyOciRepo(key string, oldValue interface{}, repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onModifyOciRepo(%s)", common.PrettyPrint(repo))
	defer log.Info("-onModifyOciRepo")

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

	ociRegistry, err := s.newOCIRegistryAndLoginWithRepo(context.Background(), repo)
	if err != nil {
		return nil, false, err
	}

	newChecksum, err := ociRegistry.checksum()
	if err != nil {
		return nil, false, err
	}

	if cacheEntry.Checksum != newChecksum {
		// TODO (gfichtenholt)
		return nil, false, status.Errorf(codes.Unimplemented, "OnModifyRepo TODO")
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
	log.Infof("+checksum()")
	defer log.Infof("-checksum()")
	appNames, err := r.listRepositoryNames()
	if err != nil {
		return "", err
	}
	tags := map[string]TagList{}
	for _, fullAppName := range appNames {
		appName, err := r.shortRepoName(fullAppName)
		if err != nil {
			return "", err
		}
		ref := fmt.Sprintf("%s/%s", r.url.String(), appName)
		tagz, err := r.getTags(ref)
		if err != nil {
			return "", err
		}

		tags[appName] = TagList{Name: appName, Tags: tagz}
	}

	content, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return common.GetSha256(content)
}

func (r *OCIRegistry) shortRepoName(fullRepoName string) (string, error) {
	expectedPrefix := strings.TrimLeft(r.url.Path, "/") + "/"
	if strings.HasPrefix(fullRepoName, expectedPrefix) {
		return fullRepoName[len(expectedPrefix):], nil
	} else {
		err := status.Errorf(codes.Internal,
			"Unexpected repository name: expected prefix: [%s], actual name: [%s]",
			expectedPrefix, fullRepoName)
		return "", err
	}
}

func (s *Server) newOCIRegistryAndLoginWithRepo(ctx context.Context, repoName types.NamespacedName) (*OCIRegistry, error) {
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	} else {
		sink := s.newRepoEventSink()
		return sink.newOCIRegistryAndLoginWithRepo(ctx, *repo)
	}
}

func (s *repoEventSink) newOCIRegistryAndLoginWithRepo(ctx context.Context, repo sourcev1.HelmRepository) (*OCIRegistry, error) {
	if loginOpts, getterOpts, cred, err := s.ociClientOptionsForRepo(ctx, repo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client: %v", err)
	} else {
		return s.newOCIRegistryAndLoginWithOptions(repo.Spec.URL, loginOpts, getterOpts, cred)
	}
}

func (s *repoEventSink) newOCIRegistryAndLoginWithOptions(registryURL string, loginOpts []registry.LoginOption, getterOpts []getter.Option, cred *orasregistryauthv2.Credential) (*OCIRegistry, error) {
	// Create registry client and login if needed.
	registryClient, file, err := newRegistryClient(loginOpts != nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client: %v", err)
	}
	if file != "" {
		defer func() {
			if err := os.Remove(file); err != nil {
				log.Infof("Failed to delete temporary credentials file: %v", err)
			}
			log.Infof("Removed temporary credentials file: [%s]", file)
		}()
	}

	registryCredentialFn := func(ctx context.Context, reg string) (orasregistryauthv2.Credential, error) {
		if cred != nil {
			return *cred, nil
		} else {
			return orasregistryauthv2.EmptyCredential, nil
		}
	}

	// a little bit misleading, since repo.Spec.URL is really an OCI Registry URL,
	// which may contain zero or more "helm repositories", such as
	// oci://demo.goharbor.io/test-oci-1, which may contain repositories "repo-1", "repo2", etc

	ociRegistry, err := newOCIRegistry(
		registryURL,
		withOCIGetter(helmGetters),
		withOCIGetterOptions(getterOpts),
		withOCIRegistryClient(registryClient),
		withRegistryCredentialFn(registryCredentialFn))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse URL '%s': %v", registryURL, err)
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

func (s *repoEventSink) ociClientOptionsForRepo(ctx context.Context, repo sourcev1.HelmRepository) ([]registry.LoginOption, []getter.Option, *orasregistryauthv2.Credential, error) {
	log.Infof("+ociClientOptionsForRepo(%s)", common.PrettyPrint(repo))
	log.Info("-ociClientOptionsForRepo")

	var loginOpts []registry.LoginOption
	var cred *orasregistryauthv2.Credential
	getterOpts := []getter.Option{
		getter.WithURL(repo.Spec.URL),
		getter.WithTimeout(repo.Spec.Timeout.Duration),
		getter.WithPassCredentialsAll(repo.Spec.PassCredentials),
	}

	secret, err := s.getRepoSecret(ctx, repo)
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
		if cred != nil {
			loginOpt := registry.LoginOptBasicAuth(cred.Username, cred.Password)
			if loginOpt != nil {
				loginOpts = append(loginOpts, loginOpt)
			}
		}
	}
	return loginOpts, getterOpts, cred, nil
}

func getOCIChartTarball(ociRegistry *OCIRegistry, chartID string, chartVersion *repo.ChartVersion) ([]byte, error) {
	chartBuffer, err := ociRegistry.downloadChart(chartVersion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return chartBuffer.Bytes(), nil
}

func getOCIChartMetadata(ociRegistry *OCIRegistry, chartID string, chartVersion *repo.ChartVersion) (*chart.Metadata, error) {
	log.Infof("+getOCIChartMetadata(%s, %s)", chartID, chartVersion.Metadata.Version)
	defer log.Infof("-getOCIChartMetadata(%s, %s)", chartID, chartVersion.Metadata.Version)

	chartTarball, err := getOCIChartTarball(ociRegistry, chartID, chartVersion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	log.Infof("==========>: chart .tgz: [%d] bytes", len(chartTarball))

	// not sure yet why flux untars into a temp directory
	files, err := tarutil.FetchChartDetailFromTarball(bytes.NewReader(chartTarball), chartID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	log.Infof("==========>: files: [%d]", len(files))

	chartYaml, ok := files[models.ChartYamlKey]
	// TODO (gfichtenholt): if there is no chart yaml (is that even possible?),
	// fall back to chart info from repo index.yaml
	if !ok || chartYaml == "" {
		return nil, status.Errorf(codes.Internal, "No chart manifest found for chart [%s]", chartID)
	}
	var chartMetadata chart.Metadata
	err = yaml.Unmarshal([]byte(chartYaml), &chartMetadata)
	if err != nil {
		return nil, err
	}
	log.Infof("==========>: chart metadata: %s", common.PrettyPrint(chartMetadata))
	return &chartMetadata, nil
}

func downloadOCIChartFn(ociRegistry *OCIRegistry) func(chartID, chartUrl, chartVersion string) ([]byte, error) {
	return func(chartID, chartUrl, chartVersion string) ([]byte, error) {
		_, chartName, err := pkgutils.SplitPackageIdentifier(chartID)
		if err != nil {
			return nil, err
		}
		cv := &repo.ChartVersion{
			URLs: []string{chartUrl},
			Metadata: &chart.Metadata{
				Name:    chartName,
				Version: chartVersion,
			},
		}
		return getOCIChartTarball(ociRegistry, chartID, cv)
	}
}
