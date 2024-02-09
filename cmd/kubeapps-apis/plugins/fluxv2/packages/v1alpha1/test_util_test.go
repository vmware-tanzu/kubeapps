// Copyright 2022-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	k8stesting "k8s.io/client-go/testing"

	helmv2beta2 "github.com/fluxcd/helm-controller/api/v2beta2"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const KubeappsCluster = "default"

type ClientReaction struct {
	verb     string
	resource string
	reaction k8stesting.ReactionFunc
}

type withWatchWrapper struct {
	delegate client.WithWatch
	watcher  *watch.RaceFreeFakeWatcher
}

func (w *withWatchWrapper) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return w.delegate.GroupVersionKindFor(obj)
}

func (w *withWatchWrapper) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return w.delegate.IsObjectNamespaced(obj)
}

var _ client.WithWatch = &withWatchWrapper{}

func (w *withWatchWrapper) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return w.delegate.Create(ctx, obj, opts...)
}

func (w *withWatchWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return w.delegate.Get(ctx, key, obj)
}

func (w *withWatchWrapper) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := w.delegate.List(ctx, list, opts...); err != nil {
		return err
	} else if accessor, err := apimeta.ListAccessor(list); err != nil {
		return err
	} else {
		accessor.SetResourceVersion("1")
		return nil
	}
}

func (w *withWatchWrapper) SubResource(subResource string) client.SubResourceClient {
	return w.delegate.SubResource(subResource)
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

func (w *withWatchWrapper) RESTMapper() apimeta.RESTMapper {
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
		return wi, fmt.Errorf("Unexpected type for watcher, expected *watch.RaceFreeFakeWatcher, got: %T", wi)
	} else {
		w.watcher = watcher
		return wi, err
	}
}

// these are helpers to compare slices ignoring order
func lessAvailablePackageFunc(p1, p2 *corev1.AvailablePackageSummary) bool {
	return p1.DisplayName < p2.DisplayName
}

// these are helpers to compare slices ignoring order
func lessInstalledPackageSummaryFunc(p1, p2 *corev1.InstalledPackageSummary) bool {
	return p1.Name < p2.Name
}

func compareJSON(t *testing.T, expectedJSON, actualJSON *apiextv1.JSON) {
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

// ref: https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
func basicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			_, err := w.Write([]byte("Unauthorised.\n"))
			if err != nil {
				log.Fatalf("%+v", err)
			}
			return
		}
		handler(w, r)
	}
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
func newBasicAuthSecret(name types.NamespacedName, user, password string) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Type: apiv1.SecretTypeOpaque,
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
func newTlsSecret(name types.NamespacedName, pub, priv, ca []byte) *apiv1.Secret {
	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
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

func newBasicAuthTlsSecret(name types.NamespacedName, user, password string, pub, priv, ca []byte) *apiv1.Secret {
	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
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
	if user != "" {
		s.Data["username"] = []byte(user)
	}
	if password != "" {
		s.Data["password"] = []byte(password)
	}
	return s
}

// ref https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
func newDockerConfigJsonSecret(name types.NamespacedName, server, user, password string) *apiv1.Secret {
	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Type: apiv1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			apiv1.DockerConfigJsonKey: []byte(`{"auths":{"` +
				server + `":{"` +
				`auth":"` + base64.StdEncoding.EncodeToString([]byte(user+":"+password)) + `"}}}`),
		},
	}
	return s
}

func setSecretManagedByKubeapps(secret *apiv1.Secret) *apiv1.Secret {
	m := secret.GetAnnotations()
	if m == nil {
		m = make(map[string]string)
		secret.SetAnnotations(m)
	}
	m[annotationManagedByKey] = annotationManagedByValue
	return secret
}

// TODO (gfichenholt) technically speaking this isn't quite right for a test case
// that actually involves non-fake k8s environment.
// In order for this to be 100% right, we need a repo object with a UID set up. But
// its good enough for a fake k8s environment, which is where this is used
func setSecretOwnerRef(repoName string, secret *apiv1.Secret) *apiv1.Secret {
	tRue := true
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         sourcev1beta2.GroupVersion.String(),
			Kind:               sourcev1beta2.HelmRepositoryKind,
			Name:               repoName,
			Controller:         &tRue,
			BlockOwnerDeletion: &tRue,
		},
	}
	return secret
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

func repoRefWithId(id string) *corev1.PackageRepositoryReference {
	// namespace will be set when scenario is run
	return repoRef(id, "TBD")
}

func repoRef(id, namespace string) *corev1.PackageRepositoryReference {
	return &corev1.PackageRepositoryReference{
		Context: &corev1.Context{
			Cluster:   KubeappsCluster,
			Namespace: namespace,
		},
		Identifier: id,
		Plugin:     fluxPlugin,
	}
}

func newCtrlClient(repos []sourcev1beta2.HelmRepository, charts []sourcev1beta2.HelmChart, releases []helmv2beta2.HelmRelease) withWatchWrapper {
	// register the flux GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	err := sourcev1beta2.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}
	err = helmv2beta2.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}

	rm := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{sourcev1beta2.GroupVersion, helmv2beta2.GroupVersion})
	rm.Add(schema.GroupVersionKind{
		Group:   sourcev1beta2.GroupVersion.Group,
		Version: sourcev1beta2.GroupVersion.Version,
		Kind:    sourcev1beta2.HelmRepositoryKind},
		apimeta.RESTScopeNamespace)
	rm.Add(schema.GroupVersionKind{
		Group:   sourcev1beta2.GroupVersion.Group,
		Version: sourcev1beta2.GroupVersion.Version,
		Kind:    sourcev1beta2.HelmChartKind},
		apimeta.RESTScopeNamespace)
	rm.Add(schema.GroupVersionKind{
		Group:   helmv2beta2.GroupVersion.Group,
		Version: helmv2beta2.GroupVersion.Version,
		Kind:    helmv2beta2.HelmReleaseKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	initLists := []client.ObjectList{}
	if len(repos) > 0 {
		initLists = append(initLists, &sourcev1beta2.HelmRepositoryList{Items: repos})
	}
	if len(charts) > 0 {
		initLists = append(initLists, &sourcev1beta2.HelmChartList{Items: charts})
	}
	if len(releases) > 0 {
		initLists = append(initLists, &helmv2beta2.HelmReleaseList{Items: releases})
	}
	if len(initLists) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(initLists...)
	}
	return withWatchWrapper{delegate: ctrlClientBuilder.Build()}
}

func ctrlClientAndWatcher(t *testing.T, s *Server) (client.WithWatch, *watch.RaceFreeFakeWatcher, error) {
	if ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster); err != nil {
		return nil, nil, err
	} else if ww, ok := ctrlClient.(*withWatchWrapper); !ok {
		return nil, nil, fmt.Errorf("Could not cast %T to: *withWatchWrapper", ctrlClient)
	} else if watcher := ww.watcher; watcher == nil {
		return nil, nil, fmt.Errorf("Unexpected condition watcher is nil")
	} else {
		return ctrlClient, watcher, nil
	}
}

func testTgz(name string) string {
	return "./testdata/charts/" + name
}

func testYaml(name string) string {
	return "./testdata/charts/" + name
}

func testCert(name string) string {
	return "./testdata/cert/" + name
}

func compareAvailablePackageDetail(t *testing.T, actual *corev1.AvailablePackageDetail, expected *corev1.AvailablePackageDetail) {
	opt1 := cmpopts.IgnoreUnexported(
		corev1.AvailablePackageDetail{},
		corev1.AvailablePackageReference{},
		corev1.Context{},
		corev1.Maintainer{},
		plugins.Plugin{},
		corev1.PackageAppVersion{})
	// these few fields a bit special in that they are all very long strings,
	// so we'll do a 'Contains' check for these instead of 'Equals'
	opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "Readme", "DefaultValues", "ValuesSchema")
	if !cmp.Equal(actual, expected, opt1, opt2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(actual, expected, opt1, opt2))
	}
	if !strings.Contains(actual.Readme, expected.Readme) {
		t.Fatalf("substring mismatch (-want: %s\n+got: %s):\n", expected.Readme, actual.Readme)
	}
	if !strings.Contains(actual.DefaultValues, expected.DefaultValues) {
		t.Fatalf("substring mismatch (-want: %s\n+got: %s):\n", expected.DefaultValues, actual.DefaultValues)
	}
	if !strings.Contains(actual.ValuesSchema, expected.ValuesSchema) {
		t.Fatalf("substring mismatch (-want: %s\n+got: %s):\n", expected.ValuesSchema, actual.ValuesSchema)
	}
}

func comparePackageRepositorySummaries(t *testing.T, actual *corev1.GetPackageRepositorySummariesResponse, expected *corev1.GetPackageRepositorySummariesResponse) {
	opts := cmpopts.IgnoreUnexported(
		corev1.Context{},
		corev1.PackageRepositoryReference{},
		plugins.Plugin{},
		corev1.PackageRepositoryStatus{},
		corev1.GetPackageRepositorySummariesResponse{},
		corev1.PackageRepositorySummary{},
	)

	// will compare this separately below
	opts2 := cmpopts.IgnoreFields(corev1.PackageRepositoryStatus{}, "UserReason")

	// cannot simply use cmpopts.SortSlices() due to doing a custom comparison of the UserReason field below.
	// Also, we don't want side effects from in-line sorting so we make a copies and use it for comparison
	// (same thing that cmp.Equal() does when you use cmpopts.SortSlices() option)
	copyA := make([]*corev1.PackageRepositorySummary, len(actual.PackageRepositorySummaries))
	copy(copyA, actual.PackageRepositorySummaries)
	sort.Slice(copyA, func(i, j int) bool { return copyA[i].Name < copyA[j].Name })

	copyE := make([]*corev1.PackageRepositorySummary, len(expected.PackageRepositorySummaries))
	copy(copyE, expected.PackageRepositorySummaries)
	sort.Slice(copyE, func(i, j int) bool { return copyE[i].Name < copyE[j].Name })

	if !cmp.Equal(copyA, copyE, opts, opts2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(copyE, copyA, opts, opts, opts2))
	}

	// now compare UserReasons, mindful of the sort order
	for i, s := range copyA {
		if !strings.HasPrefix(s.Status.UserReason, copyE[i].Status.UserReason) {
			t.Fatalf("substring mismatch (-want: %s\n+got: %s):\n",
				copyE[i].Status.UserReason,
				s.Status.UserReason)
		}
	}
}

func comparePackageRepositoryDetail(t *testing.T, actual *corev1.GetPackageRepositoryDetailResponse, expected *corev1.GetPackageRepositoryDetailResponse) {
	opts1 := cmpopts.IgnoreUnexported(
		corev1.Context{},
		corev1.PackageRepositoryReference{},
		plugins.Plugin{},
		corev1.GetPackageRepositoryDetailResponse{},
		corev1.PackageRepositoryDetail{},
		corev1.PackageRepositoryStatus{},
		corev1.PackageRepositoryAuth{},
		corev1.PackageRepositoryTlsConfig{},
		corev1.SecretKeyReference{},
		corev1.UsernamePassword{},
		corev1.TlsCertKey{},
		corev1.DockerCredentials{},
	)

	opts2 := cmpopts.IgnoreFields(corev1.PackageRepositoryStatus{}, "UserReason")

	if !cmp.Equal(expected, actual, opts1, opts2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opts1, opts2))
	}

	if !strings.HasPrefix(actual.GetDetail().Status.UserReason, expected.Detail.Status.UserReason) {
		t.Fatalf("unexpected response (status.UserReason): (-want +got):\n- %s\n+ %s",
			expected.Detail.Status.UserReason,
			actual.GetDetail().Status.UserReason)
	}
}

func compareInstalledPackageDetail(t *testing.T, actual *corev1.GetInstalledPackageDetailResponse, expected *corev1.GetInstalledPackageDetailResponse) {
	opts := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageDetailResponse{},
		corev1.InstalledPackageDetail{},
		corev1.InstalledPackageReference{},
		corev1.Context{},
		corev1.VersionReference{},
		corev1.InstalledPackageStatus{},
		corev1.PackageAppVersion{},
		plugins.Plugin{},
		corev1.ReconciliationOptions{},
		corev1.AvailablePackageReference{})
	// see comment in release_integration_test.go. Intermittently we get an inconsistent error message from flux
	opts2 := cmpopts.IgnoreFields(corev1.InstalledPackageStatus{}, "UserReason")

	// Values Applied are JSON string and need to be compared as such
	opts3 := cmpopts.IgnoreFields(corev1.InstalledPackageDetail{}, "ValuesApplied")
	if !cmp.Equal(expected, actual, opts, opts2, opts3) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opts, opts2, opts3))
	}
	if !strings.Contains(actual.InstalledPackageDetail.Status.UserReason, expected.InstalledPackageDetail.Status.UserReason) {
		t.Fatalf("substring mismatch (-want: %s\n+got: %s):\n", expected.InstalledPackageDetail.Status.UserReason, actual.InstalledPackageDetail.Status.UserReason)
	}
	compareJSONStrings(t, expected.InstalledPackageDetail.ValuesApplied, actual.InstalledPackageDetail.ValuesApplied)
}

func compareAvailablePackageSummaries(t *testing.T, actual *corev1.GetAvailablePackageSummariesResponse, expected *corev1.GetAvailablePackageSummariesResponse) {
	opt1 := cmpopts.IgnoreUnexported(
		corev1.GetAvailablePackageSummariesResponse{},
		corev1.AvailablePackageSummary{},
		corev1.AvailablePackageReference{},
		corev1.Context{},
		plugins.Plugin{},
		corev1.PackageAppVersion{})
	opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)

	if !cmp.Equal(actual, expected, opt1, opt2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opt1, opt2))
	}
}

func compareAvailablePackageVersions(t *testing.T, actual *corev1.GetAvailablePackageVersionsResponse, expected *corev1.GetAvailablePackageVersionsResponse) {
	opts := cmpopts.IgnoreUnexported(
		corev1.GetAvailablePackageVersionsResponse{},
		corev1.PackageAppVersion{})
	if !cmp.Equal(expected, actual, opts) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opts))
	}
}

func compareInstalledPackageSummaries(t *testing.T, actual *corev1.GetInstalledPackageSummariesResponse, expected *corev1.GetInstalledPackageSummariesResponse) {
	opts := cmpopts.SortSlices(lessInstalledPackageSummaryFunc)

	opts2 := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageSummariesResponse{},
		corev1.InstalledPackageSummary{},
		corev1.InstalledPackageReference{},
		corev1.InstalledPackageStatus{},
		plugins.Plugin{},
		corev1.VersionReference{},
		corev1.PackageAppVersion{},
		corev1.Context{})

	if !cmp.Equal(expected, actual, opts, opts2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opts, opts2))
	}
}

// misc global vars that get re-used in multiple tests
var (
	fluxPlugin            = &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"}
	fluxHelmRepositoryCRD = &apiextv1.CustomResourceDefinition{
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
					Status: apiextv1.ConditionStatus(metav1.ConditionTrue),
				},
			},
			StoredVersions: []string{"v1beta2"},
		},
	}
)
