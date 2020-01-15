package helm3to2

import (
	"errors"
	"net/http/httptest"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/common/response"
	h3chart "helm.sh/helm/v3/pkg/chart"
	h3 "helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	h2chart "k8s.io/helm/pkg/proto/hapi/chart"
	h2 "k8s.io/helm/pkg/proto/hapi/release"

	"testing"
)

func TestConvert(t *testing.T) {
	const (
		validSeconds   = 1452902400
		invalidSeconds = 253402300801
	)
	var (
		validDeletedTime   = helmtime.Unix(validSeconds, 0)
		invalidDeletedTime = helmtime.Unix(invalidSeconds, 0)
	)

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
		ExpectedError       error
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
		{
			Description:         "returns an error rather than panicking for a Helm 3 Release without Info",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name:  "Incomplete",
				Chart: &h3chart.Chart{},
			},
			ExpectedError: ErrUnableToConvertWithoutInfo,
		},
		{
			Description:         "returns an error for a Helm 3 Release without Chart",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name: "Incomplete",
				Info: &h3.Info{},
			},
			ExpectedError: ErrUnableToConvertWithoutInfo,
		},
		{
			Description:         "returns an error for a Helm 3 Release without Chart.Metadata",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name:  "Incomplete",
				Info:  &h3.Info{},
				Chart: &h3chart.Chart{},
			},
			ExpectedError: ErrUnableToConvertWithoutInfo,
		},
		{
			Description:         "parses and includes the deleted time",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name: "Foo",
				Info: &h3.Info{
					Status:  h3.StatusDeployed,
					Deleted: validDeletedTime,
				},
				Chart: &h3chart.Chart{
					Metadata: &h3chart.Metadata{},
					Values:   map[string]interface{}{},
				},
			},
			Helm2Release: h2.Release{
				Name: "Foo",
				Info: &h2.Info{
					Status: &h2.Status{
						Code: h2.Status_DEPLOYED,
					},
					Deleted: &timestamp.Timestamp{Seconds: validSeconds},
				},
				Chart: &h2chart.Chart{
					Metadata: &h2chart.Metadata{},
					Values: &h2chart.Config{
						Raw: "{}\n",
					},
				},
				Config: &h2chart.Config{
					Raw: "{}\n",
				},
			},
		},
		{
			Description:         "returns an error if the deleted time cannot be parsed",
			MarshallingFunction: asResponse,
			Helm3Release: h3.Release{
				Name: "Foo",
				Info: &h3.Info{
					Status:  h3.StatusDeployed,
					Deleted: invalidDeletedTime,
				},
				Chart: &h3chart.Chart{
					Metadata: &h3chart.Metadata{},
				},
			},
			ExpectedError: ErrFailedToParseDeletionTime,
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			// Perform conversion
			compatibleH3rls, err := Convert(test.Helm3Release)
			if got, want := err, test.ExpectedError; !errors.Is(got, want) {
				t.Errorf("got: %v, want: %v", got, want)
			}
			if err != nil {
				return
			}
			// Marshall both: Compatible H3Release and H2Release
			h3Marshalled := test.MarshallingFunction(compatibleH3rls)
			t.Logf("Marshalled Helm 3 Release %s", h3Marshalled)
			h2Marshalled := test.MarshallingFunction(test.Helm2Release)
			t.Logf("Marshalled Helm 2 Release %s", h2Marshalled)
			// Check result
			if h3Marshalled != h2Marshalled {
				t.Errorf("Not equal:\nMarshalled diff: %s\nUnMarshalled diff: %s",
					cmp.Diff(h3Marshalled, h2Marshalled), cmp.Diff(compatibleH3rls, test.Helm2Release))
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
