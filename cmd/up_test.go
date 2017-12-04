package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	restclient "k8s.io/client-go/rest"
)

func TestIsGKE(t *testing.T) {
	gkeVersion := version.Info{
		Major:      "foo",
		Minor:      "bar",
		GitVersion: "foobar-gke",
	}

	nonGkeVersion := version.Info{
		Major:      "foo",
		Minor:      "bar",
		GitVersion: "baz",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		output, err := json.Marshal(gkeVersion)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	defer server.Close()
	client := discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})

	if ok, err := isGKE(client); err != nil {
		t.Error(err)
	} else if !ok {
		t.Errorf("expect GKE but got non-GKE")
	}

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		output, err := json.Marshal(nonGkeVersion)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))
	defer server.Close()
	client = discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL})

	if ok, err := isGKE(client); err != nil {
		t.Error(err)
	} else if ok {
		t.Errorf("expect non-GKE but got GKE")
	}
}

func TestPrintOutput(t *testing.T) {
	om1 := metav1.ObjectMeta{
		Name:      "foo",
		Namespace: "myns",
		Labels: map[string]string{
			"created-by": "kubeapps",
		},
	}
	om2 := metav1.ObjectMeta{
		Name:      "bar",
		Namespace: "myns",
	}

	po1 := &v1.Pod{
		ObjectMeta: om1,
	}
	po2 := &v1.Pod{
		ObjectMeta: om2,
	}

	svc1 := &v1.Service{
		ObjectMeta: om1,
	}
	svc2 := &v1.Service{
		ObjectMeta: om2,
	}

	replicas := int32(1)
	sts1 := &v1beta1.StatefulSet{
		ObjectMeta: om1,
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &replicas,
		},
	}
	sts2 := &v1beta1.StatefulSet{
		ObjectMeta: om2,
	}

	dep1 := &v1beta1.Deployment{
		ObjectMeta: om1,
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
		},
	}
	dep2 := &v1beta1.Deployment{
		ObjectMeta: om2,
	}

	client := fake.NewSimpleClientset(po1, po2, svc1, svc2, sts1, sts2, dep1, dep2)
	var buf bytes.Buffer

	err := printPod(&buf, client)
	if err != nil {
		t.Error(err)
	}
	output := buf.String()
	if !strings.Contains(output, "foo") {
		t.Errorf("pod %s isn't listed", po1.Name)
	}
	if strings.Contains(output, "bar") {
		t.Errorf("pod %s shouldn't be listed", po2.Name)
	}

	err = printSvc(&buf, client)
	if err != nil {
		t.Error(err)
	}
	output = buf.String()
	if !strings.Contains(output, "foo") {
		t.Errorf("service %s isn't listed", po1.Name)
	}
	if strings.Contains(output, "bar") {
		t.Errorf("service %s shouldn't be listed", po2.Name)
	}

	err = printDeployment(&buf, client)
	if err != nil {
		t.Error(err)
	}
	output = buf.String()
	if !strings.Contains(output, "foo") {
		t.Errorf("deployment %s isn't listed", po1.Name)
	}
	if strings.Contains(output, "bar") {
		t.Errorf("deployment %s shouldn't be listed", po2.Name)
	}

	err = printStS(&buf, client)
	if err != nil {
		t.Error(err)
	}
	output = buf.String()
	if !strings.Contains(output, "foo") {
		t.Errorf("statefulset %s isn't listed", po1.Name)
	}
	if strings.Contains(output, "bar") {
		t.Errorf("statefulset %s shouldn't be listed", po2.Name)
	}
}

func TestBuildSecret(t *testing.T) {
	pw := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	name := "foo"
	ns := "my-ns"

	sr := buildSecretObject(pw, name, ns)
	if sr.Object["kind"] != "Secret" {
		t.Errorf("expect kind = secret, got %v", sr.Object["kind"])
	}
	if sr.Object["data"] == nil {
		t.Errorf("data can't be nil")
	}

	meta := sr.Object["metadata"].(map[string]interface{})
	if meta["name"].(string) != name || meta["namespace"].(string) != ns {
		t.Errorf("wrong metadata")
	}
	data := sr.Object["data"].(map[string]string)
	if data["foo"] != "bar" || data["bar"] != "baz" {
		t.Errorf("wrong data")
	}
}
func TestDump(t *testing.T) {
	objs := []*unstructured.Unstructured{}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"foo": "bar",
			"baz": map[string]string{
				"1": "2",
			},
		},
	}
	objs = append(objs, obj)
	var buf bytes.Buffer

	err := dump(&buf, "yaml", objs)
	if err != nil {
		t.Error(err)
	}
	output := buf.String()
	t.Log("output is", output)
	if !strings.Contains(output, "foo") || !strings.Contains(output, "baz") {
		t.Errorf("manifest output didn't mention both keys")
	}
}

func TestAllReady(t *testing.T) {
	po1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "myns",
			Labels: map[string]string{
				"created-by": "kubeapps",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	po2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "myns",
			Labels: map[string]string{
				"created-by": "kubeapps",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
	client := fake.NewSimpleClientset(po1, po2)
	if ok, err := allReady(client); err != nil {
		t.Error(err)
	} else if ok {
		t.Errorf("pod bar is not ready yet")
	}

	po2.Status.Phase = v1.PodRunning
	client = fake.NewSimpleClientset(po1, po2)
	if ok, err := allReady(client); err != nil {
		t.Error(err)
	} else if !ok {
		t.Errorf("expected all pods are ready, got not ready")
	}
}
