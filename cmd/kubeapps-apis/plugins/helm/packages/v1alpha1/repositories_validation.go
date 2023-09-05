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
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalog_client"
	log "k8s.io/klog/v2"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	// TODO(minelson): refactor these utils into shareable lib.
	utils "github.com/vmware-tanzu/kubeapps/cmd/asset-syncer/server"
	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

/**
  Most of the code in this file is legacy from kube_handler.go
*/

const OCIImageManifestMediaType = "application/vnd.oci.image.manifest.v1+json"
const OCIDistributionAPIProto = "https"

// ValidateRepository Checks that successful connection can be made to repository
func (s *Server) ValidateRepository(ctx context.Context, appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) error {
	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && appRepo.Namespace == s.GetGlobalPackagingNamespace() {
		// TODO(mnelson): we may also want to validate that any docker registry secrets listed
		// already exist in the namespace.
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("The docker registry secrets cannot be set for app repositories available in all namespaces"))
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
			AppRepo:             appRepo,
			Secret:              secret,
			ClientGetter:        s.repoClientGetter,
			OCICatalogAddr:      s.OCICatalogAddr,
			OCIReplacementProto: OCIDistributionAPIProto,
		}, nil
	} else {
		return HelmNonOCIValidator{
			AppRepo:      appRepo,
			Secret:       secret,
			ClientGetter: s.repoClientGetter,
		}, nil
	}
}

func newRepositoryClient(appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) (*http.Client, error) {
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
func getOCIAppRepositoryTag(cli *http.Client, repoURL string, repoName string) (string, error) {
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
		return "", fmt.Errorf("unexpected status code when querying %q: %d", repoName, resp.StatusCode)
	}

	var body []byte
	var repoTagsData repoTagsList

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error from io.ReadAll: unable to get: %v", err)
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
func getOCIAppRepositoryMediaType(client *http.Client, repoURL string, repoName string, tagVersion string) (string, error) {
	// This function is the implementation of below curl command
	// curl -XGET -H "Authorization: Basic $harborauthz"
	//		 -H "Accept: application/vnd.oci.image.manifest.v1+json"
	//		-s https://demo.goharbor.io/v2/test10/podinfo/podinfo/manifests/6.0.0

	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = path.Join("v2", parsedURL.Path, repoName, "manifests", tagVersion)

	log.InfoS("The parsedURL", "URL", parsedURL.String())
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

// validateOCIAppRepository validates OCI Repos only
// return true if mediaType == "application/vnd.cncf.helm.config" otherwise false
func (v *HelmOCIValidator) validateOCIAppRepository(ctx context.Context, appRepo *apprepov1alpha1.AppRepository) (bool, error) {

	repoURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/")
	// If the app repo url was specified using the oci protocol - "oci://" -
	// then we need to replace it to interact with the http(s) distribution
	// spec API.
	repoURL = strings.Replace(repoURL, "oci://", fmt.Sprintf("%s://", v.OCIReplacementProto), 1)

	var httpCLI *http.Client
	var err error
	if v.ClientGetter != nil {
		httpCLI, err = v.ClientGetter(v.AppRepo, v.Secret)
		if err != nil {
			return false, err
		}
	}

	// For the OCI case, if no repositories are listed then we validate that a
	// catalog is available for the registry, otherwise we want to validate that
	// all the listed repositories are valid
	if len(appRepo.Spec.OCIRepositories) == 0 {
		if v.ClientGetter == nil {
			return false, fmt.Errorf("an http ClientGetter is required to validate a repository")
		}
		u, err := url.Parse(repoURL)
		if err != nil {
			log.Errorf("Could not parse URL: %+v", err)
			return false, err
		}
		var grpcClient ocicatalog.OCICatalogServiceClient
		if v.OCICatalogAddr != "" {
			var closer func()
			grpcClient, closer, err = ocicatalog_client.NewClient(v.OCICatalogAddr)
			if err != nil {
				return false, err
			}
			defer closer()
		}
		oci := utils.OciAPIClient{RegistryNamespaceUrl: u, HttpClient: httpCLI, GrpcClient: grpcClient}
		if available, err := oci.CatalogAvailable(ctx, ""); err != nil {
			log.Errorf("error calling CatalogAvailable: %+v", err)
			return false, err
		} else if available {
			return true, nil
		}

		return false, ErrEmptyOCIRegistry
	}

	for _, repoName := range appRepo.Spec.OCIRepositories {
		tagVersion, err := getOCIAppRepositoryTag(httpCLI, repoURL, repoName)
		if err != nil {
			return false, err
		}

		mediaType, err := getOCIAppRepositoryMediaType(httpCLI, repoURL, repoName, tagVersion)
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
			return nil, fmt.Errorf("unable to parse validation response. Got: %w", err)
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
	// OCIReplacementProto is only a field on the struct so that in tests we
	// can use `http` rather than `https`. When a HelmOCIValidator is created
	// here in non-test code, we use `https`.
	OCIReplacementProto string
}

func (v HelmOCIValidator) Validate(ctx context.Context) (*ValidationResponse, error) {
	// We need to either have an http client getter or access
	// to the OCI Catalog service.
	if v.OCICatalogAddr == "" && v.ClientGetter == nil {
		return nil, fmt.Errorf("unable to validate without either http client or OCI Catalog address")
	}

	response := &ValidationResponse{Code: 200, Message: "OK"}
	isValidRepo, err := v.validateOCIAppRepository(ctx, v.AppRepo)
	if err != nil || !isValidRepo {
		response = &ValidationResponse{Code: 400, Message: err.Error()}
	}
	return response, nil
}
