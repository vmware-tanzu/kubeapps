package helm3to2

import (
	"net/http/httptest"
	"runtime/debug"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/common/response"
	h3chart "helm.sh/helm/v3/pkg/chart"
	h3 "helm.sh/helm/v3/pkg/release"
	h2chart "k8s.io/helm/pkg/proto/hapi/chart"
	h2 "k8s.io/helm/pkg/proto/hapi/release"

	"testing"
)

func TestConvert(t *testing.T) {
	type testScenario struct {
		// Scenario params
		Description  string
		Helm3Release h3.Release
		Helm2Release h2.Release
		// MarshallingFunction defines what it means to
		// parse a Helm 2 Release to a string i.e. which fields are relevant/redundant,
		// and how to format the fields of a Helm 2 release.
		// E.g. spew.Dumps is a valid marhsalling function.
		MarshallingFunction func(h2.Release) string
	}
	tests := []testScenario{
		{
			Description:         "Two equivalent releases",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name:      "Foo",
				Namespace: "default",
				Chart: &h3chart.Chart{
					Metadata: &h3chart.Metadata{},
					Values: map[string]interface{}{
						"port": 8080,
					},
				},
				Info: &h3.Info{
					Status: h3.StatusDeployed,
				},
				Version: 1,
				Config: map[string]interface{}{
					"port": 3000,
					"user": map[string]interface{}{
						"name":     "user1",
						"password": "123456",
					},
				},
			},
			Helm2Release: h2.Release{
				Name:      "Foo",
				Namespace: "default",
				Info: &h2.Info{
					Status: &h2.Status{
						Code: h2.Status_DEPLOYED,
					},
				},
				Chart: &h2chart.Chart{
					Metadata: &h2chart.Metadata{},
					Values: &h2chart.Config{
						Raw: "port: 8080\n",
					},
				},
				Version: 1,
				Config: &h2chart.Config{
					Raw: "port: 3000\nuser:\n  name: user1\n  password: \"123456\"\n",
				},
			},
		},
		{
			Description:         "Two equivalent releases with switched order of values",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name:      "Foo",
				Namespace: "default",
				Chart: &h3chart.Chart{
					Metadata: &h3chart.Metadata{},
					Values:   map[string]interface{}{},
				},
				Info: &h3.Info{
					Status: h3.StatusDeployed,
				},
				Version: 1,
				Config: map[string]interface{}{
					"user": map[string]interface{}{
						"password": "123456",
						"name":     "user1",
					},
					"port": 3000,
				},
			},
			Helm2Release: h2.Release{
				Name:      "Foo",
				Namespace: "default",
				Info: &h2.Info{
					Status: &h2.Status{
						Code: h2.Status_DEPLOYED,
					},
				},
				Chart: &h2chart.Chart{
					Metadata: &h2chart.Metadata{},
					Values: &h2chart.Config{
						Raw: "{}\n",
					},
				},
				Version: 1,
				Config: &h2chart.Config{
					Raw: "port: 3000\nuser:\n  name: user1\n  password: \"123456\"\n",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			// Capture the panic and report it in an orderly fashion
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Got a panic: %v. \nStacktrace: \n%s", r, string(debug.Stack()))
				}
			}()
			// Perform conversion
			compatibleH3rls, err := Convert(test.Helm3Release)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			// Marshall both: Compatible H3Release and H2Release
			h3Marshalled := test.MarshallingFunction(compatibleH3rls)
			t.Logf("Marshalled Helm 3 Release %s", h3Marshalled)
			h2Marshalled := test.MarshallingFunction(test.Helm2Release)
			t.Logf("Marshalled Helm 2 Release %s", h2Marshalled)
			// Check result
			if h3Marshalled != h2Marshalled {
				t.Errorf("Not equal: %s", cmp.Diff(h3Marshalled, h2Marshalled))
			}
		})
	}
}

// asResponse is one of many possible Marshalling functions
// However, it's the one most relevant for the current usage of 'dashboardcompat.go'
func asResponse(data h2.Release) string {
	w := httptest.NewRecorder()
	response.NewDataResponse(data).Write(w)
	return w.Body.String()
}
