package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	k8scorev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const KubeappsCluster = "default"

type withWatchWrapper struct {
	delegate client.WithWatch
	watcher  *watch.RaceFreeFakeWatcher
}

var _ client.WithWatch = &withWatchWrapper{}

func (w *withWatchWrapper) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return w.delegate.Create(ctx, obj, opts...)
}

func (w *withWatchWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return w.delegate.Get(ctx, key, obj)
}

func (w *withWatchWrapper) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := w.delegate.List(ctx, list, opts...); err != nil {
		return err
	} else if accessor, err := meta.ListAccessor(list); err != nil {
		return err
	} else {
		accessor.SetResourceVersion("1")
		return nil
	}
}

func (w *withWatchWrapper) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return w.delegate.Delete(ctx, obj, opts...)
}

func (w *withWatchWrapper) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return w.delegate.DeleteAllOf(ctx, obj, opts...)
}

func (w *withWatchWrapper) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return w.delegate.Patch(ctx, obj, patch, opts...)
}

func (w *withWatchWrapper) RESTMapper() meta.RESTMapper {
	return w.delegate.RESTMapper()
}

func (w *withWatchWrapper) Scheme() *runtime.Scheme {
	return w.delegate.Scheme()
}

func (w *withWatchWrapper) Status() client.StatusWriter {
	return w.delegate.Status()
}

func (w *withWatchWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return w.delegate.Update(ctx, obj, opts...)
}

func (w *withWatchWrapper) Watch(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
	wi, err := w.delegate.Watch(ctx, list, opts...)
	if err != nil {
		return wi, err
	} else if watcher, ok := wi.(*watch.RaceFreeFakeWatcher); !ok {
		return wi, fmt.Errorf(
			"Unexpected type for watcher, expected *watch.RaceFreeFakeWatcher, got: %s",
			reflect.TypeOf(wi))
	} else {
		w.watcher = watcher
		return wi, err
	}
}

// these are helpers to compare slices ignoring order
func lessAvailablePackageFunc(p1, p2 *corev1.AvailablePackageSummary) bool {
	return p1.DisplayName < p2.DisplayName
}

func lessPackageRepositoryFunc(p1, p2 *v1alpha1.PackageRepository) bool {
	return p1.Name < p2.Name && p1.Namespace < p2.Namespace
}

// these are helpers to compare slices ignoring order
func lessInstalledPackageSummaryFunc(p1, p2 *corev1.InstalledPackageSummary) bool {
	return p1.Name < p2.Name
}

func compareJSON(t *testing.T, expectedJSON, actualJSON *v1.JSON) {
	expectedJSONString, actualJSONString := "", ""
	if expectedJSON != nil {
		expectedJSONString = string(expectedJSON.Raw)
	}
	if actualJSON != nil {
		actualJSONString = string(actualJSON.Raw)
	}
	compareJSONStrings(t, expectedJSONString, actualJSONString)
}

func compareJSONStrings(t *testing.T, expectedJSONString, actualJSONString string) {
	var expected interface{}
	if expectedJSONString != "" {
		if err := json.Unmarshal([]byte(expectedJSONString), &expected); err != nil {
			t.Fatal(err)
		}
	}
	var actual interface{}
	if actualJSONString != "" {
		if err := json.Unmarshal([]byte(actualJSONString), &actual); err != nil {
			t.Fatal(err)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(actual); err != nil {
			t.Fatal(err)
		}
		if expected != buf.String() {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(expectedJSONString, buf.String()))
		}
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
func newBasicAuthSecret(name, namespace, user, password string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: k8scorev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(password),
		},
	}
}

// Note that according to https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
// TLS secrets need to look one way, but according to
// https://fluxcd.io/docs/components/source/helmrepositories/#spec-examples they expect TLS secrets
// in a different format:
// certFile/keyFile/caFile vs tls.crt/tls.key. I am going with flux's example for now:
func newTlsSecret(name, namespace string, pub, priv, ca []byte) (*k8scorev1.Secret, error) {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: k8scorev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"certFile": pub,
			"keyFile":  priv,
			"caFile":   ca,
		},
	}, nil
}

func newBasicAuthTlsSecret(name, namespace, user, password string, pub, priv, ca []byte) (*k8scorev1.Secret, error) {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: k8scorev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(password),
			"certFile": pub,
			"keyFile":  priv,
			"caFile":   ca,
		},
	}, nil
}

func availableRef(id, namespace string) *corev1.AvailablePackageReference {
	return &corev1.AvailablePackageReference{
		Identifier: id,
		Context: &corev1.Context{
			Namespace: namespace,
			Cluster:   KubeappsCluster,
		},
		Plugin: fluxPlugin,
	}
}

func installedRef(id, namespace string) *corev1.InstalledPackageReference {
	return &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: namespace,
			Cluster:   KubeappsCluster,
		},
		Identifier: id,
		Plugin:     fluxPlugin,
	}
}

// misc global vars that get re-used in multiple tests
var fluxPlugin = &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"}
var fluxHelmRepositoryCRD = &apiextv1.CustomResourceDefinition{
	TypeMeta: metav1.TypeMeta{
		Kind:       "CustomResourceDefinition",
		APIVersion: "apiextensions.k8s.io/v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "helmrepositories.source.toolkit.fluxcd.io",
	},
	Status: apiextv1.CustomResourceDefinitionStatus{
		Conditions: []apiextv1.CustomResourceDefinitionCondition{
			{
				Type:   "Established",
				Status: "True",
			},
		},
	},
}
