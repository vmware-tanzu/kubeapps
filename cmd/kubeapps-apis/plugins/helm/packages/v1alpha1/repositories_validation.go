// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	"google.golang.org/grpc"
	log "k8s.io/klog/v2"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	// TODO(minelson): refactor these utils into shareable lib.
	utils "github.com/vmware-tanzu/kubeapps/cmd/asset-syncer/server"
	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
)

/**
  Most of the code in this file is legacy from kube_handler.go
*/

const OCIImageManifestMediaType = "application/vnd.oci.image.manifest.v1+json"

// ValidateRepository Checks that successful connection can be made to repository
func (s *Server) ValidateRepository(ctx context.Context, appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) error {
	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && appRepo.Namespace == s.GetGlobalPackagingNamespace() {
		// TODO(mnelson): we may also want to validate that any docker registry secrets listed
		// already exist in the namespace.
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("docker registry secrets cannot be set for app repositories available in all namespaces"))
	}

	validator, err := s.getValidator(appRepo, secret)
	if err != nil {
		return err
	}
	resp, err := validator.Validate(ctx)
	if err != nil {
		return err
	} else if resp.Code >= 400 {
		log.Errorf("Failed repository validation validation: %+v", resp)
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Failed repository validation: %v", resp))
	} else {
		return nil
	}
}

// getValidator return appropriate RepositoryValidator interface for OCI and
// non-OCI Repos
func (s *Server) getValidator(appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) (RepositoryValidator, error) {
	if appRepo.Spec.Type == "oci" {
		// For the OCI case, we want to validate that all the given repositories are valid
		return HelmOCIValidator{
			AppRepo:        appRepo,
			Secret:         secret,
			ClientGetter:   s.repoClientGetter,
			OCICatalogAddr: s.OCICatalogAddr,
		}, nil
	} else {
		return HelmNonOCIValidator{
			AppRepo:      appRepo,
			Secret:       secret,
			ClientGetter: s.repoClientGetter,
		}, nil
	}
}

func newRepositoryClient(appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) (httpclient.Client, error) {
	if cli, err := helm.InitNetClient(appRepo, secret, secret, nil); err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Unable to create HTTP client for repository: %w", err))
	} else {
		return cli, nil
	}
}

// ValidationResponse represents the response after validating a repo
type ValidationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrEmptyOCIRegistry defines the error returned when an attempt is
// made to create an OCI registry with no repositories
var ErrEmptyOCIRegistry = fmt.Errorf("unable to determine the OCI catalog, you need to specify at least one repository")

// repoTagsList stores the list of tags for an OCI repository.
type repoTagsList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type repoConfig struct {
	MediaType string `json:"mediaType"`
}

// repoManifest stores the mediatype for an OCI repository.
type repoManifest struct {
	Config repoConfig `json:"config"`
}

// getOCIAppRepositoryTag Gets a tag for the given repo URL & name
func getOCIAppRepositoryTag(cli httpclient.Client, repoURL string, repoName string) (string, error) {
	// This function is the implementation of below curl command
	// curl -XGET -H "Authorization: Basic $harborauthz"
	//		-H "Accept: application/vnd.oci.image.manifest.v1+json"
	//		-s https://demo.goharbor.io/v2/test10/podinfo/podinfo/tags/list\?n\=1

	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path = path.Join("v2", parsedURL.Path, repoName, "tags", "list")
	q := parsedURL.Query()
	q.Add("n", "1")
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	//This header is required for a successful request
	req.Header.Set("Accept", OCIImageManifestMediaType)

	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected status code when querying %q: %d", repoName, resp.StatusCode)
	}

	var body []byte
	var repoTagsData repoTagsList

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("io.ReadAll : unable to get: %v", err)
		return "", err
	}

	err = json.Unmarshal(body, &repoTagsData)
	if err != nil {
		err = fmt.Errorf("OCI Repo tag at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}

	if len(repoTagsData.Tags) == 0 {
		err = fmt.Errorf("OCI Repo tag at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}

	tagVersion := repoTagsData.Tags[0]
	return tagVersion, nil
}

// getOCIAppRepositoryMediaType Gets manifests config.MediaType for the given repo URL & Name
func getOCIAppRepositoryMediaType(client httpclient.Client, repoURL string, repoName string, tagVersion string) (string, error) {
	// This function is the implementation of below curl command
	// curl -XGET -H "Authorization: Basic $harborauthz"
	//		 -H "Accept: application/vnd.oci.image.manifest.v1+json"
	//		-s https://demo.goharbor.io/v2/test10/podinfo/podinfo/manifests/6.0.0

	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = path.Join("v2", parsedURL.Path, repoName, "manifests", tagVersion)

	log.InfoS("parsedURL", "URL", parsedURL.String())
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	//This header is required for a successful request
	req.Header.Set("Accept", OCIImageManifestMediaType)

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mediaData repoManifest

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &mediaData)
	if err != nil {
		err = fmt.Errorf("OCI Repo manifest at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}
	mediaType := mediaData.Config.MediaType
	return mediaType, nil
}

// ValidateOCIAppRepository validates OCI Repos only
// return true if mediaType == "application/vnd.cncf.helm.config" otherwise false
func ValidateOCIAppRepository(appRepo *apprepov1alpha1.AppRepository, cli httpclient.Client) (bool, error) {

	repoURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/")
	// For the OCI case, if no repositories are listed then we validate that a
	// catalog is available for the registry, otherwise we want to validate that
	// all the listed repositories are valid
	if len(appRepo.Spec.OCIRepositories) == 0 {
		u, err := url.Parse(repoURL)
		if err != nil {
			log.Errorf("could not parse URL: %+v", err)
			return false, err
		}
		oci := utils.OciAPIClient{AuthHeader: "", Url: u, NetClient: cli}
		if oci.CatalogAvailable("") {
			return true, nil
		}
		return false, ErrEmptyOCIRegistry
	}
	for _, repoName := range appRepo.Spec.OCIRepositories {
		tagVersion, err := getOCIAppRepositoryTag(cli, repoURL, repoName)
		if err != nil {
			return false, err
		}

		mediaType, err := getOCIAppRepositoryMediaType(cli, repoURL, repoName, tagVersion)
		if err != nil {
			return false, err
		}

		if !strings.HasPrefix(mediaType, "application/vnd.cncf.helm.config") {
			err := fmt.Errorf("%v is not a Helm OCI Repo. mediaType starting with %q expected, found %q", repoName, "application/vnd.cncf.helm.config", mediaType)
			return false, err
		}
	}
	return true, nil
}

// RepositoryValidator is an interface for checking the validity of an
// AppRepository
type RepositoryValidator interface {
	// Validate returns a validation response.
	Validate(context.Context) (*ValidationResponse, error)
}

// HelmNonOCIValidator is an HttpValidator for non-OCI Helm repositories.
type HelmNonOCIValidator struct {
	AppRepo      *apprepov1alpha1.AppRepository
	Secret       *corev1.Secret
	ClientGetter repositoryClientGetter
}

func (r HelmNonOCIValidator) Validate(ctx context.Context) (*ValidationResponse, error) {
	repoURL := strings.TrimSuffix(strings.TrimSpace(r.AppRepo.Spec.URL), "/")
	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}
	cli, err := r.ClientGetter(r.AppRepo, r.Secret)
	if err != nil {
		return nil, err
	}

	res, err := cli.Do(req)
	if err != nil {
		// If the request fail, it's not an internal error
		return &ValidationResponse{Code: 400, Message: err.Error()}, nil
	}
	response := &ValidationResponse{Code: res.StatusCode, Message: "OK"}
	if response.Code != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse validation response. Got: %w", err)
		}
		response.Message = string(body)
	}

	return response, nil
}

type HelmOCIValidator struct {
	AppRepo        *apprepov1alpha1.AppRepository
	Secret         *corev1.Secret
	ClientGetter   repositoryClientGetter
	OCICatalogAddr string
}

func queryOCICatalog(ctx context.Context, ociCatalogAddr string, appRepoURL string) (*ValidationResponse, error) {
	u, err := url.Parse(appRepoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL for OCI Catalog %q: %+v", appRepoURL, err)
	}

	conn, err := grpc.Dial(ociCatalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("unable to contact OCI Catalog at %q: %+v", ociCatalogAddr, err)
	}
	defer conn.Close()

	client := ocicatalog.NewOCICatalogServiceClient(conn)

	repos_stream, err := client.ListRepositoriesForRegistry(ctx, &ocicatalog.ListRepositoriesForRegistryRequest{
		Registry:     u.Host,
		Namespace:    u.Path,
		ContentTypes: []string{"helm"},
	})
	if err != nil {
		return nil, fmt.Errorf("error querying OCI Catalog for repos: %+v", err)
	}

	// It's enough to receive a single repo to be valid.
	_, err = repos_stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("error receiving OCI Repositories: %+v", err)
	}
	return &ValidationResponse{Code: 200, Message: "OK"}, nil
}

func (v HelmOCIValidator) Validate(ctx context.Context) (*ValidationResponse, error) {
	// We need to either have an http client getter or access
	// to the OCI Catalog service.
	if v.OCICatalogAddr == "" && v.ClientGetter == nil {
		return nil, fmt.Errorf("unable to validate without either http client or OCI Catalog address")
	}

	// Prefer the OCI Catalog service, but we just log and ignore errors
	// for now so that behaviour does not change for VAC index support.
	if v.OCICatalogAddr != "" {
		resp, err := queryOCICatalog(ctx, v.OCICatalogAddr, v.AppRepo.Spec.URL)
		if err == nil {
			return resp, nil
		}
		log.Errorf("unable to query OCI Catalog service at %q: %+v", v.OCICatalogAddr, err)
	}

	if v.ClientGetter == nil {
		return nil, fmt.Errorf("unable to validate without http client")
	}
	var response *ValidationResponse
	response = &ValidationResponse{Code: 200, Message: "OK"}

	cli, err := v.ClientGetter(v.AppRepo, v.Secret)
	if err != nil {
		return nil, err
	}
	// If there was an error validating the OCI repository, it's not an internal error.
	isValidRepo, err := ValidateOCIAppRepository(v.AppRepo, cli)
	if err != nil || !isValidRepo {
		response = &ValidationResponse{Code: 400, Message: err.Error()}
	}
	return response, nil
}
