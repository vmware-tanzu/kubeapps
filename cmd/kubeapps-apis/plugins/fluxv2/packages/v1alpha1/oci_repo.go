// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Inspired by
// https://github.com/fluxcd/source-controller/blob/main/internal/helm/repository/
//  oci_chart_repository.go
// and adapted for kubeapps use
// OCI spec ref
// https://github.com/opencontainers/distribution-spec/blob/main/spec.md

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/Masterminds/semver/v3"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	log "k8s.io/klog/v2"

	"github.com/fluxcd/pkg/version"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	//	"github.com/fluxcd/source-controller/internal/transport"
)

// RegistryClient is an interface for interacting with OCI registries
// It is used by the OCIChartRepository to retrieve chart versions
// from OCI registries
type RegistryClient interface {
	Login(host string, opts ...registry.LoginOption) error
	Logout(host string, opts ...registry.LogoutOption) error
	Tags(url string) ([]string, error)
}

// OCIChartRepository represents a Helm chart repository, and the configuration
// required to download the repository tags and charts from the repository.
// All methods are thread safe unless defined otherwise.
type OCIChartRepository struct {
	// URL is the location of the repository.
	URL url.URL
	// Client to use while accessing the repository's contents.
	Client getter.Getter
	// Options to configure the Client with while downloading tags
	// or a chart from the URL.
	Options []getter.Option

	tlsConfig *tls.Config

	// RegistryClient is a client to use while downloading tags or charts from a registry.
	RegistryClient RegistryClient
}

// OCIChartRepositoryOption is a function that can be passed to NewOCIChartRepository
// to configure an OCIChartRepository.
type OCIChartRepositoryOption func(*OCIChartRepository) error

// WithOCIRegistryClient returns a ChartRepositoryOption that will set the registry client
func WithOCIRegistryClient(client RegistryClient) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.RegistryClient = client
		return nil
	}
}

// WithOCIGetter returns a ChartRepositoryOption that will set the getter.Getter
func WithOCIGetter(providers getter.Providers) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		c, err := providers.ByScheme(r.URL.Scheme)
		if err != nil {
			return err
		}
		r.Client = c
		return nil
	}
}

// WithOCIGetterOptions returns a ChartRepositoryOption that will set the getter.Options
func WithOCIGetterOptions(getterOpts []getter.Option) OCIChartRepositoryOption {
	return func(r *OCIChartRepository) error {
		r.Options = getterOpts
		return nil
	}
}

// NewOCIChartRepository constructs and returns a new ChartRepository with
// the ChartRepository.Client configured to the getter.Getter for the
// repository URL scheme. It returns an error on URL parsing failures.
// It assumes that the url scheme has been validated to be an OCI scheme.
func NewOCIChartRepository(repositoryURL string, chartRepoOpts ...OCIChartRepositoryOption) (*OCIChartRepository, error) {
	u, err := url.Parse(repositoryURL)
	if err != nil {
		return nil, err
	}

	r := &OCIChartRepository{}
	r.URL = *u
	for _, opt := range chartRepoOpts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Get returns the repo.ChartVersion for the given name, the version is expected
// to be a semver.Constraints compatible string. If version is empty, the latest
// stable version will be returned and prerelease versions will be ignored.
// adapted from https://github.com/helm/helm/blob/49819b4ef782e80b0c7f78c30bd76b51ebb56dc8/pkg/downloader/chart_downloader.go#L162
func (r *OCIChartRepository) Get(name, ver string) (*repo.ChartVersion, error) {
	// Find chart versions matching the given name.
	// Either in an index file or from a registry.
	ref := fmt.Sprintf("%s/%s", r.URL.String(), name)
	log.Infof("about to call getTags(%s)", ref)
	cvs, err := r.getTags(ref)
	if err != nil {
		return nil, err
	}

	if len(cvs) == 0 {
		return nil, fmt.Errorf("unable to locate any tags in provided repository: %s", name)
	}

	// Determine if version provided
	// If empty, try to get the highest available tag
	// If exact version, try to find it
	// If semver constraint string, try to find a match
	tag, err := getLastMatchingVersionOrConstraint(cvs, ver)
	return &repo.ChartVersion{
		URLs: []string{fmt.Sprintf("%s/%s:%s", r.URL.String(), name, tag)},
		Metadata: &chart.Metadata{
			Name:    name,
			Version: tag,
		},
	}, err
}

// This function shall be called for OCI registries only
// It assumes that the ref has been validated to be an OCI reference.
func (r *OCIChartRepository) getTags(ref string) ([]string, error) {
	url := strings.TrimPrefix(ref, fmt.Sprintf("%s://", registry.OCIScheme))
	log.Infof("about to call RegistryClient.Tags(%s)", url)
	// Retrieve list of repository tags
	tags, err := r.RegistryClient.Tags(url)
	log.Infof("done with call RegistryClient.Tags(%s): %s %v", url, tags, err)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("unable to locate any tags in provided repository: %s", ref)
	}

	return tags, nil
}

// DownloadChart confirms the given repo.ChartVersion has a downloadable URL,
// and then attempts to download the chart using the Client and Options of the
// ChartRepository. It returns a bytes.Buffer containing the chart data.
// In case of an OCI hosted chart, this function assumes that the chartVersion url is valid.
func (r *OCIChartRepository) DownloadChart(chart *repo.ChartVersion) (*bytes.Buffer, error) {
	if len(chart.URLs) == 0 {
		return nil, fmt.Errorf("chart '%s' has no downloadable URLs", chart.Name)
	}

	ref := chart.URLs[0]
	u, err := url.Parse(ref)
	if err != nil {
		err = fmt.Errorf("invalid chart URL format '%s': %w", ref, err)
		return nil, err
	}

	ustr := strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	return nil, status.Errorf(codes.Unimplemented, "TODO %s", ustr)

	/* TODO:
	t := transport.NewOrIdle(r.tlsConfig)
	clientOpts := append(r.Options, getter.WithTransport(t))
	defer transport.Release(t)

	// trim the oci scheme prefix if needed
	return r.Client.Get(ustr, clientOpts...)
	*/
}

// Login attempts to login to the OCI registry.
// It returns an error on failure.
func (r *OCIChartRepository) Login(opts ...registry.LoginOption) error {
	err := r.RegistryClient.Login(r.URL.Host, opts...)
	if err != nil {
		return err
	}
	return nil
}

// Logout attempts to logout from the OCI registry.
// It returns an error on failure.
func (r *OCIChartRepository) Logout() error {
	err := r.RegistryClient.Logout(r.URL.Host)
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

// NewRegistryClient generates a registry client and a temporary credential file.
// The client is meant to be used for a single reconciliation.
// The file is meant to be used for a single reconciliation and deleted after.
func NewRegistryClient(isLogin bool) (*registry.Client, string, error) {
	if isLogin {
		// create a temporary file to store the credentials
		// this is needed because otherwise the credentials are stored in ~/.docker/config.json.
		credentialFile, err := os.CreateTemp("", "credentials")
		if err != nil {
			return nil, "", err
		}

		rClient, err := registry.NewClient(registry.ClientOptWriter(io.Discard), registry.ClientOptCredentialsFile(credentialFile.Name()))
		if err != nil {
			return nil, "", err
		}
		return rClient, credentialFile.Name(), nil
	}

	rClient, err := registry.NewClient(registry.ClientOptWriter(io.Discard))
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

	// Construct the Getter options from the HelmRepository data
	loginOpts, getterOpts, err := s.helmOptionsForRepo(repo)
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "failed to create registry client options: %v", err)
	}
	log.Infof("=====================> loginOpts: [%v], getterOpts: [%v]", loginOpts, getterOpts)

	// Create registry client and login if needed.
	registryClient, file, err := NewRegistryClient(loginOpts != nil)
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "failed to create registry client: %v", err)
	}
	if file != "" {
		defer func() {
			if err := os.Remove(file); err != nil {
				log.Infof("Failed to delete temporary credentials file: %v", err)
			}
		}()
	}

	chartRepo, err := NewOCIChartRepository(
		repo.Spec.URL,
		WithOCIRegistryClient(registryClient))
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "failed to parse URL '%s': %v", repo.Spec.URL, err)
	}

	// Attempt to login to the registry if credentials are provided.
	if loginOpts != nil {
		err = chartRepo.Login(loginOpts...)
		if err != nil {
			return nil, false, err
		}
	}

	repoURL, err := url.ParseRequestURI(strings.TrimSpace(repo.Spec.URL))
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "%v", err)
	}

	// e.g.
	log.Infof("==========>: url: [%s]", common.PrettyPrint(repoURL))

	ref := "podinfo"
	log.Infof("==========>: ref: [%s]", ref)

	chartVersion, err := chartRepo.Get(ref, "")
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "%v", err)
	}
	log.Infof("==========>: chart version: %s", common.PrettyPrint(chartVersion))

	// later we'll do
	// chartRepo.DownloadChart(chartVersion)
	// to cache the charts' .tgz file

	return nil, false, nil
	/*
			authorizationHeader := ""
		// TODO look at repo's secretRef
		/*
			// The auth header may be a dockerconfig that we need to parse
			if serveOpts.DockerConfigJson != "" {
				dockerConfig := &credentialprovider.DockerConfigJSON{}
				err := json.Unmarshal([]byte(serveOpts.DockerConfigJson), dockerConfig)
				if err != nil {
					return nil, false, status.Errorf(codes.Internal, "%v", err)
				}
				authorizationHeader, err = kube.GetAuthHeaderFromDockerConfig(dockerConfig)
				if err != nil {
					return nil, false, status.Errorf(codes.Internal, "%v", err)
				}
			}
		headers := http.Header{}
		if authorizationHeader != "" {
			headers["Authorization"] = []string{authorizationHeader}
		}
		// TODO look at repo's secretRef for TLS config
		netClient, err := httpclient.NewWithCertFile(additionalCAFile, false)
		if err != nil {
			return nil, false, status.Errorf(codes.Internal, "%v", err)
		}

		ociResolver :=
		docker.NewResolver(
			docker.ResolverOptions{
				Headers: headers,

				Client:  netClient})
			// from cmd/asset-syncer/server/utils.go
			//func pullAndExtract(repoURL *url.URL, appName, tag string, puller helm.ChartPuller, r *OCIRegistry) (*models.Chart, error) {
			repoURL, err := url.ParseRequestURI(strings.TrimSpace(repo.Spec.URL))
			if err != nil {
				return nil, false, status.Errorf(codes.Internal, "%v", err)
			}
			repositories := []string{"test-oci-1/podinfo"}

			// find tags
			// see func (r *OCIRegistry) FilterIndex()
			// TagList represents a list of tags as specified at
			// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#content-discovery
			type TagList struct {
				Name string   `json:"name"`
				Tags []string `json:"tags"`
			}
			tags := map[string]TagList{}

			for _, appName := range repositories {
				tag := tags[appName].Tags[0]
				ref := path.Join(repoURL.Host, repoURL.Path, fmt.Sprintf("%s:%s", appName, tag))
			}

				chartBuffer, digest, err := puller.PullOCIChart(ref)
				if err != nil {
					return nil, err
				}

				// Extract
				files, err := extractFilesFromBuffer(chartBuffer)
				if err != nil {
					return nil, err
				}
				chartMetadata := chart.Metadata{}
				err = yaml.Unmarshal([]byte(files.Metadata), &chartMetadata)
				if err != nil {
					return nil, err
				}

				// Format Data
				chartVersion := models.ChartVersion{
					Version:    chartMetadata.Version,
					AppVersion: chartMetadata.AppVersion,
					Digest:     digest,
					URLs:       chartMetadata.Sources,
					Readme:     files.Readme,
					Values:     files.Values,
					Schema:     files.Schema,
				}

				maintainers := []chart.Maintainer{}
				for _, m := range chartMetadata.Maintainers {
					maintainers = append(maintainers, chart.Maintainer{
						Name:  m.Name,
						Email: m.Email,
						URL:   m.URL,
					})
				}

				// Encode repository names to store them in the database.
				encodedAppName := url.PathEscape(appName)

					&models.Chart{
						ID:            path.Join(r.Name, encodedAppName),
						Name:          encodedAppName,
						Repo:          &models.Repo{Namespace: r.Namespace, Name: r.Name, URL: r.URL, Type: r.Type},
						Description:   chartMetadata.Description,
						Home:          chartMetadata.Home,
						Keywords:      chartMetadata.Keywords,
						Maintainers:   maintainers,
						Sources:       chartMetadata.Sources,
						Icon:          chartMetadata.Icon,
						Category:      chartMetadata.Annotations["category"],
						ChartVersions: []models.ChartVersion{chartVersion},
					}, nil
	*/
}

//
// misc OCI repo utilities
//
func (s *repoEventSink) helmOptionsForRepo(repo sourcev1.HelmRepository) ([]registry.LoginOption, []getter.Option, error) {
	log.Infof("+helmOptionsForRepo()")

	getterOpts := []getter.Option{
		getter.WithURL(repo.Spec.URL),
		getter.WithTimeout(repo.Spec.Timeout.Duration),
		getter.WithPassCredentialsAll(repo.Spec.PassCredentials),
	}

	secret, err := s.getRepoSecret(context.Background(), repo)
	if err != nil {
		return nil, nil, err
	} else if secret != nil {
		opts, err := common.HelmGetterOptionsFromSecret(*secret)
		if err != nil {
			return nil, nil, err
		}
		getterOpts = append(getterOpts, opts...)
	}

	log.Infof("============> getter opts: [%v]", getterOpts)

	/*
		ctx :=
		_, err := s.clientOptionsForRepo(ctx, repo)
		if err != nil {
			return nil, nil, err
		}

				loginOpt := append(getterOpts, common.ConvertClientOptionsToHelmGetterOptions(c)...)

					tlsConfig, err = getter.TLSClientConfigFromSecret(*secret, repo.Spec.URL)
					if err != nil {
						return nil, err
					}

					// Build registryClient options from secret
					loginOpt, err := registry.LoginOptionFromSecret(repo.Spec.URL, *secret)
					if err != nil {
						return nil, err
					}

			loginOpts := append([]registry.LoginOption{}, loginOpt)
	*/
	return nil, getterOpts, nil
}
