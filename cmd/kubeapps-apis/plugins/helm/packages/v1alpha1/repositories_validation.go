// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
)

/**
  Most of the code in this file is legacy from kube_handler.go
*/

// ValidateRepository Checks that successful connection can be made to repository
func (s *Server) ValidateRepository(appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) error {
	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && appRepo.Namespace == s.GetGlobalPackagingNamespace() {
		// TODO(mnelson): we may also want to validate that any docker registry secrets listed
		// already exist in the namespace.
		return status.Errorf(codes.FailedPrecondition, "docker registry secrets cannot be set for app repositories available in all namespaces")
	}

	client, err := s.repoClientGetter(appRepo, secret)
	if err != nil {
		return err
	}
	httpValidator, err := getValidator(appRepo)
	if err != nil {
		return err
	}
	resp, err := httpValidator.Validate(client)
	if err != nil {
		return err
	} else if resp.Code >= 400 {
		return status.Errorf(codes.FailedPrecondition, "Failed repository validation: %v", resp)
	} else {
		return nil
	}
}

// getValidator return appropriate HttpValidator interface for OCI and non-OCI Repos
func getValidator(appRepo *apprepov1alpha1.AppRepository) (kube.HttpValidator, error) {
	if appRepo.Spec.Type == "oci" {
		// For the OCI case, we want to validate that all the given repositories are valid
		if len(appRepo.Spec.OCIRepositories) == 0 {
			return nil, kube.ErrEmptyOCIRegistry
		}
		return kube.HelmOCIValidator{
			AppRepo: appRepo,
		}, nil
	} else {
		repoURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/")
		parsedURL, err := url.ParseRequestURI(repoURL)
		if err != nil {
			return nil, err
		}
		parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return nil, err
		}
		return kube.HelmNonOCIValidator{
			Req: req,
		}, nil
	}
}

func newRepositoryClient(appRepo *apprepov1alpha1.AppRepository, secret *corev1.Secret) (httpclient.Client, error) {
	if cli, err := kube.InitNetClient(appRepo, secret, secret, nil); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Unable to create HTTP client for repository: %v", err)
	} else {
		return cli, nil
	}
}
