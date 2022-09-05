// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	KubeappsCluster      = "default"
	AppRepositoryGroup   = "kubeapps.com"
	AppRepositoryVersion = "v1alpha1"
	AppRepositoryApi     = AppRepositoryGroup + "/" + AppRepositoryVersion
)

// misc global vars that get re-used in multiple tests
var (
	helmPlugin           = &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"}
	helmAppRepositoryCRD = &apiextv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apprepositories.kubeapps.com",
		},
		Status: apiextv1.CustomResourceDefinitionStatus{
			Conditions: []apiextv1.CustomResourceDefinitionCondition{
				{
					Type:   "Established",
					Status: apiextv1.ConditionStatus(metav1.ConditionTrue),
				},
			},
			StoredVersions: []string{"v1beta2"},
		},
	}
)

func repoRef(id, cluster, namespace string) *corev1.PackageRepositoryReference {
	return &corev1.PackageRepositoryReference{
		Context: &corev1.Context{
			Cluster:   cluster,
			Namespace: namespace,
		},
		Identifier: id,
		Plugin:     helmPlugin,
	}
}

// these are helpers to compare slices ignoring order
func lessPackageRepositorySummaryFunc(p1, p2 *corev1.PackageRepositorySummary) bool {
	return p1.Name < p2.Name
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
func newBasicAuthSecret(name, namespace, username, password string) *apiv1.Secret {
	authString := fmt.Sprintf("%s:%s", username, password)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	return newAuthTokenSecret(name, namespace, authHeader)
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
func newAuthTokenSecret(name, namespace, token string) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{
			"authorizationHeader": []byte(token),
		},
	}
}

func newAuthDockerSecret(name, namespace, jsonData string) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: apiv1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": []byte(jsonData),
		},
	}
}

func dockerAuthJson(server, username, password, email, auth string) string {
	return fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"%s\",\"auth\":\"%s\"}}}",
		server, username, password, email, auth)
}

func setSecretOwnerRef(repoName string, secret *apiv1.Secret) *apiv1.Secret {
	tRue := true
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         AppRepositoryApi,
			Kind:               AppRepositoryKind,
			Name:               repoName,
			Controller:         &tRue,
			BlockOwnerDeletion: &tRue,
		},
	}
	return secret
}

// Note that according to https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
// TLS secrets need to look one way, but according to
func newTlsSecret(name, namespace string, pub, priv, ca []byte) *apiv1.Secret {
	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
	if pub != nil {
		s.Data["certFile"] = pub
	}
	if priv != nil {
		s.Data["keyFile"] = priv
	}
	if ca != nil {
		s.Data["ca.crt"] = ca
	}
	return s
}

func newRepoHttpClient(responses map[string]*http.Response) newRepoClient {
	return func(appRepo *appRepov1.AppRepository, secret *apiv1.Secret) (httpclient.Client, error) {
		return &fakeHTTPClient{
			responses: responses,
		}, nil
	}
}

type fakeHTTPClient struct {
	responses map[string]*http.Response
}

func (f *fakeHTTPClient) Do(h *http.Request) (*http.Response, error) {
	if resp, ok := f.responses[h.URL.String()]; !ok {
		return nil, fmt.Errorf("url requested '%s' not found in valid responses %v", h.URL.String(), f.responses)
	} else {
		return resp, nil
	}
}

func httpResponse(statusCode int, body string) *http.Response {
	return &http.Response{StatusCode: statusCode, Body: io.NopCloser(bytes.NewReader([]byte(body)))}
}

func testCert(name string) string {
	return "./testdata/cert/" + name
}

func testYaml(name string) string {
	return "./testdata/charts/" + name
}

// generate-cert.sh script in testdata directory is used to generate these files
func getCertsForTesting(t *testing.T) (ca, pub, priv []byte) {
	var err error
	if ca, err = os.ReadFile(testCert("ca.pem")); err != nil {
		t.Fatalf("%+v", err)
	} else if pub, err = os.ReadFile(testCert("server.pem")); err != nil {
		t.Fatalf("%+v", err)
	} else if priv, err = os.ReadFile(testCert("server-key.pem")); err != nil {
		t.Fatalf("%+v", err)
	}
	return ca, pub, priv
}

func newCtrlClient(repos []*appRepov1.AppRepository) client.WithWatch {
	// Register required schema definitions
	scheme := runtime.NewScheme()
	err := appRepov1.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("%s", err)
	}

	rm := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{{Group: AppRepositoryGroup, Version: AppRepositoryVersion}})
	rm.Add(schema.GroupVersionKind{
		Group:   AppRepositoryGroup,
		Version: AppRepositoryVersion,
		Kind:    AppRepositoryKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	var initLists []client.ObjectList
	if len(repos) > 0 {
		repoInst := make([]appRepov1.AppRepository, len(repos))
		for i, repo := range repos {
			repoInst[i] = *repo
		}
		initLists = append(initLists, &appRepov1.AppRepositoryList{Items: repoInst})
	}
	if len(initLists) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(initLists...)
	}
	return ctrlClientBuilder.Build()
}
