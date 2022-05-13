package main

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"io/ioutil"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	KubeappsCluster      = "default"
	AppRepositoryGroup   = "kubeapps.com"
	AppRepositoryVersion = "v1alpha1"
	AppRepositoryApi     = AppRepositoryGroup + "/" + AppRepositoryVersion
	AppRepositoryKind    = "AppRepository"
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

func repoRef(id, namespace string) *corev1.PackageRepositoryReference {
	return &corev1.PackageRepositoryReference{
		Context: &corev1.Context{
			Cluster:   KubeappsCluster,
			Namespace: namespace,
		},
		Identifier: id,
		Plugin:     helmPlugin,
	}
}

// ref: https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
func basicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
func newBasicAuthSecret(name, namespace, username, password string) *apiv1.Secret {
	authString := fmt.Sprintf("%s:%s", username, password)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{
			"authenticationHeader": []byte(authHeader),
		},
	}
}

func tlsAuth(pub, priv []byte) *corev1.PackageRepositoryAuth {
	return &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
		PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_TlsCertKey{
			TlsCertKey: &corev1.TlsCertKey{
				Cert: string(pub),
				Key:  string(priv),
			},
		},
	}
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
// https://fluxcd.io/docs/components/source/helmrepositories/#spec-examples they expect TLS secrets
// in a different format:
// certFile/keyFile/caFile vs tls.crt/tls.key. I am going with flux's example for now:
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
		s.Data["caFile"] = ca
	}
	return s
}

func testCert(name string) string {
	return "./testdata/cert/" + name
}

// generate-cert.sh script in testdata directory is used to generate these files
func getCertsForTesting(t *testing.T) (ca, pub, priv []byte) {
	var err error
	if ca, err = ioutil.ReadFile(testCert("ca.pem")); err != nil {
		t.Fatalf("%+v", err)
	} else if pub, err = ioutil.ReadFile(testCert("server.pem")); err != nil {
		t.Fatalf("%+v", err)
	} else if priv, err = ioutil.ReadFile(testCert("server-key.pem")); err != nil {
		t.Fatalf("%+v", err)
	}
	return ca, pub, priv
}

func newCtrlClient(repos []appRepov1.AppRepository) client.WithWatch {
	// Register required schema definitions
	scheme := runtime.NewScheme()
	appRepov1.AddToScheme(scheme)

	rm := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{{AppRepositoryGroup, AppRepositoryVersion}})
	rm.Add(schema.GroupVersionKind{
		Group:   AppRepositoryGroup,
		Version: AppRepositoryVersion,
		Kind:    AppRepositoryKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	var initLists []client.ObjectList
	if len(repos) > 0 {
		initLists = append(initLists, &appRepov1.AppRepositoryList{Items: repos})
	}
	if len(initLists) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(initLists...)
	}
	return ctrlClientBuilder.Build()
}
