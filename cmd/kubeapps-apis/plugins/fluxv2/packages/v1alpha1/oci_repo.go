// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Inspired by
// https://github.com/fluxcd/source-controller/blob/main/internal/helm/repository/oci_chart_repository.go
// and adapted for kubeapps use. Why is there a need for this? There are features flux chose not to address,
// such as
//   - listing of available repositories for a given registry
//     ref https://github.com/fluxcd/source-controller/issues/839
//   - listing of available charts for a given repository
// while kubeapps needs those features in order to support plugin features like
// GetAvailablePackageSummaries/GetAvailablePackageVersions, etc.
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
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common/transport"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"

	"github.com/fluxcd/pkg/oci/auth/login"
	"github.com/fluxcd/pkg/version"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"

	// OCI Registry As a Storage (ORAS)
	orasregistryauthv2 "oras.land/oras-go/v2/registry/remote/auth"
)

// RegistryClient is an interface for interacting with OCI registries
// It is used by the OCIChartRepository to retrieve chart versions
// from OCI registries. Functions Login/Logout/Tags are implemented by
// https://github.com/helm/helm/blob/main/pkg/registry/client.go
// DownloadChart is implemented below
type RegistryClientDownloadChartFn func(*repo.ChartVersion) (*bytes.Buffer, error)
type RegistryClient interface {
	Login(host string, opts ...registry.LoginOption) error
	Logout(host string, opts ...registry.LogoutOption) error
	Tags(url string) ([]string, error)
	DownloadChart(chartVersion *repo.ChartVersion) (*bytes.Buffer, error)
}

// an interface flux plugin uses to determine what kind of vendor-specific
// registry repository name lister applies, and then executes specific logic
type OCIChartRepositoryLister interface {
	IsApplicableFor(*OCIChartRepository) (bool, error)
	ListRepositoryNames(OCIChartRepository *OCIChartRepository) ([]string, error)
}

// OCIChartRepository represents a Helm chart repository, and the configuration
// required to download the repository tags and charts from the repository.
// All methods are thread safe unless defined otherwise.
type OCIChartRepository struct {
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
	registryCredentialFn OCIChartRepositoryCredentialFn

	repositoryLister OCIChartRepositoryLister
}

// OCIChartRepositoryOption is a function that can be passed to newOCIChartRepository()
// to configure an OCIChartRepository.
type OCIChartRepositoryOption func(*OCIChartRepository) error

type OCIChartRepositoryCredentialFn func(ctx context.Context, reg string) (orasregistryauthv2.Credential, error)

var (
	helmProviders = getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
		getter.Provider{
			Schemes: []string{"oci"},
			New:     getter.NewOCIGetter,
		},
	}

	// TODO (gfichtenholt) make this extensible so code coming from other plugs/modules
	// can register new repository listers
	builtInRepoListers = []OCIChartRepositoryLister{
		NewDockerRegistryApiV2RepositoryLister(),
		NewHarborRegistryApiV2RepositoryLister(),
		// TODO (gfichtenholt) other container registry providers, like AWS, Azure, etc
	}

	// the reason for so many arguments to this func, as opposed to an OCIChartRepository instance is
	// the result of this func is used to build a new instance of OCIChartRepository, i.e. a avoiding
	// catch-22
	registryClientBuilderFn = func(isLogin bool, tlsConfig *tls.Config, getterOpts []getter.Option, helmGetter getter.Getter) (RegistryClient, string, error) {
		return newRegistryClient(isLogin, tlsConfig, getterOpts, helmGetter)
	}
)

type registryClientType struct {
	registryClient  *registry.Client
	chartDownloader RegistryClientDownloadChartFn
}

func (c *registryClientType) Login(host string, opts ...registry.LoginOption) error {
	return c.registryClient.Login(host, opts...)
}

func (c *registryClientType) Logout(host string, opts ...registry.LogoutOption) error {
	return c.registryClient.Logout(host, opts...)
}

func (c *registryClientType) Tags(url string) ([]string, error) {
	return c.registryClient.Tags(url)
}

func (c *registryClientType) DownloadChart(chartVersion *repo.ChartVersion) (*bytes.Buffer, error) {
	return c.chartDownloader(chartVersion)
}

// withRegistryClient returns a OCIChartRepositoryOption that will set the registry client
func withRegistryClient(client RegistryClient) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.registryClient = client
		return nil
	}
}

// withHelmGetter returns a OCIChartRepositoryOption that will set the getter.Getter
func withHelmGetter(providers getter.Providers) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		c, err := providers.ByScheme(r.url.Scheme)
		if err != nil {
			return err
		}
		r.helmGetter = c
		return nil
	}
}

// withHelmGetterOptions returns a OCIChartRepositoryOption that will set the getter.Options
func withHelmGetterOptions(getterOpts []getter.Option) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.helmOptions = getterOpts
		return nil
	}
}

func withRegistryCredentialFn(fn OCIChartRepositoryCredentialFn) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.registryCredentialFn = fn
		return nil
	}
}

func withTlsConfig(c *tls.Config) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.tlsConfig = c
		return nil
	}
}

// newOCIChartRepository constructs and returns a new OCIChartRepository with
// the RegistryClient configured to the getter.Getter for the
// registry URL scheme. It returns an error on URL parsing failures.
// It assumes that the url scheme has been validated to be an OCI scheme.
func newOCIChartRepository(registryURL string, registryOpts ...OCIChartRepositoryOption) (*OCIChartRepository, error) {
	u, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}

	r := &OCIChartRepository{}
	r.url = *u
	for _, opt := range registryOpts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *OCIChartRepository) listRepositoryNames() ([]string, error) {
	log.Infof("+listRepositoryNames")

	// this needs to be done after a call to login()
	if r.repositoryLister == nil {
		for _, lister := range builtInRepoListers {
			if ok, err := lister.IsApplicableFor(r); ok && err == nil {
				r.repositoryLister = lister
				break
			} else {
				log.Infof("Lister [%v] not applicable for registry with URL [%s] due to: [%v]",
					reflect.TypeOf(lister), r.url.String(), err)
			}
		}
	}

	if r.repositoryLister == nil {
		return nil, status.Errorf(
			codes.Internal,
			"No repository lister found for OCI registry with URL: [%s]",
			r.url.String())
	}

	return r.repositoryLister.ListRepositoryNames(r)
}

// pickChartVersionFrom returns the ChartVersion for the given name, the version is expected
// to be a semver.Constraints compatible string. If version is empty, the latest
// stable version will be returned and prerelease versions will be ignored.
// adapted from https://github.com/helm/helm/blob/49819b4ef782e80b0c7f78c30bd76b51ebb56dc8/pkg/downloader/chart_downloader.go#L162
func (r *OCIChartRepository) pickChartVersionFrom(name, ver string, cvs []string) (*repo.ChartVersion, error) {
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
func (r *OCIChartRepository) getTags(ref string) ([]string, error) {
	log.Infof("+getTags(%s)", ref)
	defer log.Infof("-getTags(%s)", ref)

	ref = strings.TrimPrefix(ref, fmt.Sprintf("%s://", registry.OCIScheme))

	log.Infof("getTags: about to call .Tags(%s)", ref)
	tags, err := r.registryClient.Tags(ref)
	log.Infof("getTags: done with call .Tags(%s): tags: %s, err: %v", ref, tags, err)
	if err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("unable to locate any tags in provided repository: %s", ref)
	}
	return tags, nil
}

// TODO (gfichtenholt) Call this at some point :)
// logout attempts to logout from the OCI registry.
//nolint:unused
func (r *OCIChartRepository) logout() error {
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
func newHelmRegistryClient(isLogin bool) (*registry.Client, string, error) {
	clientOpts := []registry.ClientOption{
		registry.ClientOptWriter(io.Discard),
	}

	var file string
	if isLogin {
		// create a temporary file to store the credentials
		// this is needed because otherwise the credentials are stored in ~/.docker/config.json.
		credentialFile, err := os.CreateTemp("", "credentials")
		if err != nil {
			return nil, "", err
		}
		file = credentialFile.Name()

		clientOpts = append(clientOpts,
			registry.ClientOptCredentialsFile(credentialFile.Name()),
		)
	}
	rClient, err := registry.NewClient(clientOpts...)
	return rClient, file, err
}

func newRegistryClient(isLogin bool, tlsConfig *tls.Config, getterOpts []getter.Option, helmGetter getter.Getter) (RegistryClient, string, error) {
	rClient, file, err := newHelmRegistryClient(isLogin)
	if err != nil {
		return nil, file, err
	}

	chartDownloader := func(chartVersion *repo.ChartVersion) (*bytes.Buffer, error) {
		getterOpts = append(getterOpts, getter.WithRegistryClient(rClient))
		return downloadChartWithHelmGetter(tlsConfig, getterOpts, helmGetter, chartVersion)
	}

	return &registryClientType{
		rClient, chartDownloader,
	}, file, nil
}

// OCI Helm repository, which defines a source, does not produce an Artifact
// ref https://fluxcd.io/docs/components/source/helmrepositories/#helm-oci-repository

func (s *repoEventSink) onAddOciRepo(repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onAddOciRepo(%s)", common.PrettyPrint(repo))
	defer log.Info("-onAddOciRepo")

	ociChartRepo, err := s.newOCIChartRepositoryAndLogin(context.Background(), repo)
	if err != nil {
		return nil, false, err
	}
	// repository names, e.g. "stefanprodan/charts/podinfo"
	// asset-syncer calls them appNames
	// see func (r *OCIRegistry) Charts(fetchLatestOnly bool) ([]models.Chart, error) {
	// also per https://github.com/helm/community/blob/main/hips/hip-0006.md#4-chart-names--oci-reference-basenames
	// appName == chartName == the basename (the last segment of the URL path) on a registry reference
	appNames, err := ociChartRepo.listRepositoryNames()
	if err != nil {
		return nil, false, err
	}

	allTags, err := ociChartRepo.getTagsForApps(appNames)
	if err != nil {
		return nil, false, err
	}

	charts, err := getOciChartModels(appNames, allTags, ociChartRepo, &repo)
	if err != nil {
		return nil, false, err
	}

	checksum, err := ociChartRepo.checksum(appNames, allTags)
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "%v", err)
	}

	cacheEntryValue := repoCacheEntryValue{
		Checksum: checksum,
		Charts:   charts,
		Type:     "oci",
	}

	// use gob encoding instead of json, it peforms much better
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(cacheEntryValue); err != nil {
		return nil, false, err
	}

	if s.chartCache != nil {
		fn := downloadOCIChartFn(ociChartRepo)
		if err = s.chartCache.SyncCharts(charts, fn); err != nil {
			return nil, false, err
		}
	}
	return buf.Bytes(), true, nil
}

func (s *repoEventSink) onModifyOciRepo(key string, oldValue interface{}, repo sourcev1.HelmRepository) ([]byte, bool, error) {
	log.Infof("+onModifyOciRepo(repo:%s)", common.PrettyPrint(repo))
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

	ociChartRepo, err := s.newOCIChartRepositoryAndLogin(context.Background(), repo)
	if err != nil {
		return nil, false, err
	}

	appNames, err := ociChartRepo.listRepositoryNames()
	if err != nil {
		return nil, false, err
	}

	allTags, err := ociChartRepo.getTagsForApps(appNames)
	if err != nil {
		return nil, false, err
	}

	newChecksum, err := ociChartRepo.checksum(appNames, allTags)
	if err != nil {
		return nil, false, err
	}

	if cacheEntry.Checksum != newChecksum {
		charts, err := getOciChartModels(appNames, allTags, ociChartRepo, &repo)
		if err != nil {
			return nil, false, err
		}

		cacheEntryValue := repoCacheEntryValue{
			Checksum: newChecksum,
			Charts:   charts,
			Type:     "oci",
		}

		// use gob encoding instead of json, it peforms much better
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err = enc.Encode(cacheEntryValue); err != nil {
			return nil, false, err
		}

		if s.chartCache != nil {
			fn := downloadOCIChartFn(ociChartRepo)
			if err = s.chartCache.SyncCharts(charts, fn); err != nil {
				return nil, false, err
			}
		}
		return buf.Bytes(), true, nil
	}
	return nil, false, nil
}

// TagList represents a list of tags as specified at
// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#content-discovery
type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// Checksum returns the sha256 of the repo by concatenating tags for
// all repositories within the registry and returning the sha256.
// Caveat: Mutated image tags won't be detected as new
func (r *OCIChartRepository) getTagsForApps(appNames []string) (map[string]TagList, error) {
	tags := map[string]TagList{}
	for _, fullAppName := range appNames {
		appName, err := r.shortRepoName(fullAppName)
		if err != nil {
			return nil, err
		}
		ref := fmt.Sprintf("%s/%s", r.url.String(), appName)
		tagz, err := r.getTags(ref)
		if err != nil {
			return nil, err
		}
		tags[appName] = TagList{Name: appName, Tags: tagz}
	}
	return tags, nil
}

func (r *OCIChartRepository) checksum(appNames []string, allTags map[string]TagList) (string, error) {
	log.Infof("+checksum(%s)", appNames)
	defer log.Infof("-checksum()")

	content, err := json.Marshal(allTags)
	if err != nil {
		return "", err
	}

	return common.GetSha256(content)
}

//
// misc OCI repo utilities
//

// given fullRepoName like "stefanprodan/charts/podinfo", returns "podinfo"
func (r *OCIChartRepository) shortRepoName(fullRepoName string) (string, error) {
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

func (s *Server) newOCIChartRepositoryAndLogin(ctx context.Context, repoName types.NamespacedName) (*OCIChartRepository, error) {
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	} else {
		sink := s.newRepoEventSink()
		return sink.newOCIChartRepositoryAndLogin(ctx, *repo)
	}
}

func (s *repoEventSink) newOCIChartRepositoryAndLogin(ctx context.Context, repo sourcev1.HelmRepository) (*OCIChartRepository, error) {
	if loginOpts, getterOpts, cred, err := s.clientOptionsForOciRepo(ctx, repo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client: %v", err)
	} else {
		return s.newOCIChartRepositoryAndLoginWithOptions(repo.Spec.URL, loginOpts, getterOpts, cred)
	}
}

func (s *repoEventSink) newOCIChartRepositoryAndLoginWithOptions(registryURL string, loginOpts []registry.LoginOption, getterOpts []getter.Option, cred *orasregistryauthv2.Credential) (*OCIChartRepository, error) {
	u, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}
	helmProvider, err := helmProviders.ByScheme(u.Scheme)
	if err != nil {
		return nil, err
	}

	var tlsConfig *tls.Config

	// Create new registry client and login if needed.
	registryClient, file, err := registryClientBuilderFn(loginOpts != nil, tlsConfig, getterOpts, helmProvider)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client due to: %v", err)
	}
	if registryClient == nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry client")
	}
	if file != "" {
		defer func() {
			if err := os.Remove(file); err != nil {
				log.Infof("Failed to delete temporary credentials file: %v", err)
			}
			log.Infof("Successfully removed temporary credentials file: [%s]", file)
		}()
	}

	registryCredentialFn := func(ctx context.Context, reg string) (orasregistryauthv2.Credential, error) {
		log.Infof("+ORAS registryCredentialFn(%s)", reg)
		if cred != nil {
			return *cred, nil
		} else {
			return orasregistryauthv2.EmptyCredential, nil
		}
	}

	// a little bit misleading, since repo.Spec.URL is really an OCI Registry URL,
	// which may contain zero or more "helm repositories", such as
	// oci://demo.goharbor.io/test-oci-1, which may contain repositories "repo-1", "repo2", etc
	ociRepo, err := newOCIChartRepository(
		registryURL,
		withHelmGetter(helmProviders),
		withHelmGetterOptions(getterOpts),
		withRegistryClient(registryClient),
		withRegistryCredentialFn(registryCredentialFn),
		withTlsConfig(tlsConfig))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse URL '%s': %v", registryURL, err)
	}

	// Attempt to login to the registry if credentials are provided.
	if loginOpts != nil {
		err := ociRepo.registryClient.Login(ociRepo.url.Host, loginOpts...)
		log.Infof("login(%s): %v", ociRepo.url.Host, err)
		if err != nil {
			return nil, err
		}
	}
	return ociRepo, nil
}

func (s *repoEventSink) clientOptionsForOciRepo(ctx context.Context, repo sourcev1.HelmRepository) ([]registry.LoginOption, []getter.Option, *orasregistryauthv2.Credential, error) {
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

		cred, err = common.OCIChartRepositoryCredentialFromSecret(repo.Spec.URL, *secret)
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

	if repo.Spec.Provider != "" && repo.Spec.Provider != sourcev1.GenericOCIProvider {
		ctxTimeout, cancel := context.WithTimeout(ctx, repo.Spec.Timeout.Duration)
		defer cancel()

		cred, err = oidcAuth(ctxTimeout, repo)
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

// downloadChartWithHelmGetter() confirms the given repo.ChartVersion has a downloadable URL,
// and then attempts to download the chart using the Client and Options of the
// OCIChartRepository. It returns a bytes.Buffer containing the chart data.
// In case of an OCI hosted chart, this function assumes that the chartVersion url is valid.
func downloadChartWithHelmGetter(tlsConfig *tls.Config, getterOptions []getter.Option, helmGetter getter.Getter, chartVersion *repo.ChartVersion) (*bytes.Buffer, error) {
	log.Infof("+downloadChartWithHelmGetter(%s)", chartVersion.Version)
	if len(chartVersion.URLs) == 0 {
		return nil, fmt.Errorf("chart '%s' has no downloadable URLs", chartVersion.Name)
	}

	ref := chartVersion.URLs[0]
	u, err := url.Parse(ref)
	if err != nil {
		err = fmt.Errorf("invalid chart URL format '%s': %w", ref, err)
		return nil, err
	}

	t := transport.NewOrIdle(tlsConfig)
	clientOpts := append(getterOptions, getter.WithTransport(t))

	defer func() {
		if err = transport.Release(t); err != nil {
			log.Errorf("%+v", err)
		}
	}()

	if err != nil {
		return nil, err
	}

	// trim the oci scheme prefix if needed
	getThis := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	log.Infof("about to call helmGetter.Get(%s)", getThis)
	buf, err := helmGetter.Get(getThis, clientOpts...)
	if buf != nil {
		log.Infof("helmGetter.Get(%s) returned buffer size: [%d] error: %v", getThis, len(buf.Bytes()), err)
	} else {
		log.Infof("helmGetter.Get(%s) returned error: %v", getThis, err)
	}
	return buf, err
}

func getOciChartModels(appNames []string, allTags map[string]TagList, ociChartRepo *OCIChartRepository, repo *sourcev1.HelmRepository) ([]models.Chart, error) {
	charts := []models.Chart{}
	for _, fullAppName := range appNames {
		appName, err := ociChartRepo.shortRepoName(fullAppName)
		if err != nil {
			return nil, err
		}

		tags, ok := allTags[appName]
		if !ok {
			return nil, status.Errorf(codes.Internal, "Missing tags for app [%s]", appName)
		}

		mc, err := getOciChartModel(appName, tags, ociChartRepo, repo)
		if err != nil {
			return nil, err
		}
		charts = append(charts, *mc)
	}
	return charts, nil
}

func getOciChartModel(appName string, tags TagList, ociChartRepo *OCIChartRepository, repo *sourcev1.HelmRepository) (*models.Chart, error) {
	// Encode repository names to store them in the database.
	encodedAppName := url.PathEscape(appName)
	chartID := path.Join(repo.Name, encodedAppName)

	log.Infof("==========>: app name: [%s], chartID: [%s]", appName, chartID)

	// to be consistent with how we support helm http repos
	// the chart fields like Desciption, home, sources come from the
	// most recent chart version
	// ref https://github.com/vmware-tanzu/kubeapps/blob/11c87926d6cd798af72875d01437d15ae8d85b9a/pkg/helm/index.go#L30
	latestChartVersion, err := ociChartRepo.pickChartVersionFrom(appName, "", tags.Tags)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	log.Infof("==========> most recent chart version: %s", latestChartVersion.Version)

	latestChartMetadata, err := getOCIChartMetadata(ociChartRepo, chartID, latestChartVersion)
	if err != nil {
		return nil, err
	}

	maintainers := []chart.Maintainer{}
	for _, maintainer := range latestChartMetadata.Maintainers {
		maintainers = append(maintainers, *maintainer)
	}

	modelRepo := &models.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Spec.URL,
		Type:      repo.Spec.Type,
	}

	mc := models.Chart{
		ID:            chartID,
		Name:          encodedAppName,
		Repo:          modelRepo,
		Description:   latestChartMetadata.Description,
		Home:          latestChartMetadata.Home,
		Keywords:      latestChartMetadata.Keywords,
		Maintainers:   maintainers,
		Sources:       latestChartMetadata.Sources,
		Icon:          latestChartMetadata.Icon,
		Category:      latestChartMetadata.Annotations["category"],
		ChartVersions: []models.ChartVersion{},
	}

	for _, tag := range tags.Tags {
		chartVersion, err := ociChartRepo.pickChartVersionFrom(appName, tag, tags.Tags)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%v", err)
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
	return &mc, nil
}

func getOCIChartTarball(ociRepo *OCIChartRepository, chartID string, chartVersion *repo.ChartVersion) ([]byte, error) {
	chartBuffer, err := ociRepo.registryClient.DownloadChart(chartVersion)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return chartBuffer.Bytes(), nil
}

func getOCIChartMetadata(ociRepo *OCIChartRepository, chartID string, chartVersion *repo.ChartVersion) (*chart.Metadata, error) {
	log.Infof("+getOCIChartMetadata(%s, %s)", chartID, chartVersion.Metadata.Version)
	defer log.Infof("-getOCIChartMetadata(%s, %s)", chartID, chartVersion.Metadata.Version)

	chartTarball, err := getOCIChartTarball(ociRepo, chartID, chartVersion)
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

func downloadOCIChartFn(ociRepo *OCIChartRepository) func(chartID, chartUrl, chartVersion string) ([]byte, error) {
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
		return getOCIChartTarball(ociRepo, chartID, cv)
	}
}

// oidcAuth generates the OIDC credential authenticator based on the specified cloud provider.
func oidcAuth(ctx context.Context, repo sourcev1.HelmRepository) (*orasregistryauthv2.Credential, error) {
	url := strings.TrimPrefix(repo.Spec.URL, sourcev1.OCIRepositoryPrefix)
	ref, err := name.ParseReference(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", repo.Spec.URL, err)
	}

	cred, err := loginWithManager(ctx, repo.Spec.Provider, url, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to login to registry '%s': %w", repo.Spec.URL, err)
	}

	return cred, nil
}

func loginWithManager(ctx context.Context, provider, url string, ref name.Reference) (*orasregistryauthv2.Credential, error) {
	opts := login.ProviderOptions{}
	switch provider {
	case sourcev1.AmazonOCIProvider:
		opts.AwsAutoLogin = true
	case sourcev1.AzureOCIProvider:
		opts.AzureAutoLogin = true
	case sourcev1.GoogleOCIProvider:
		opts.GcpAutoLogin = true
	}

	auth, err := login.NewManager().Login(ctx, url, ref, opts)
	if err != nil {
		return nil, err
	}

	if auth == nil {
		return nil, nil
	}

	return common.OIDCAdaptHelper(auth)
}
