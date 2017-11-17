package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
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
