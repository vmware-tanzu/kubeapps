// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/disintegration/imaging"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	helmfake "github.com/vmware-tanzu/kubeapps/pkg/helm/fake"
	helmtest "github.com/vmware-tanzu/kubeapps/pkg/helm/test"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalog_client"
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalog_client/ocicatalog_clienttest"
	tartest "github.com/vmware-tanzu/kubeapps/pkg/tarutil/test"
	"helm.sh/helm/v3/pkg/chart"
	log "k8s.io/klog/v2"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

var validRepoIndexYAMLBytes, _ = os.ReadFile("testdata/valid-index.yaml")
var validRepoIndexYAML = string(validRepoIndexYAMLBytes)

const chartsIndexManifestJSON = `
{
	"schemaVersion": 2,
	"config": {
	  "mediaType": "application/vnd.vmware.charts.index.config.v1+json",
	  "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
	  "size": 2
	},
	"layers": [
	  {
		"mediaType": "application/vnd.vmware.charts.index.layer.v1+json",
		"digest": "sha256:f9f7df0ae3f50aaf9ff390034cec4286d2aa43f061ce4bc7aa3c9ac862800aba",
		"size": 1169,
		"annotations": {
		  "org.opencontainers.image.title": "charts-index.json"
		}
	  }
	]
  }
`
const chartsIndexMultipleJSON = `
{
    "entries": {
        "common": {
            "versions": [
                {
                    "version": "2.4.0",
                    "appVersion": "2.4.0",
                    "name": "common",
                    "urls": [
                        "harbor.example.com/charts/common:2.4.0"
                    ],
                    "digest": "sha256:c85139bbe83ec5af6201fe1bec39fc0d0db475de41bc74cd729acc5af8eed6ba",
                    "releasedAt": "2023-06-08T12:15:48.149853788Z"
                }
            ]
        },
        "redis": {
            "versions": [
                {
                    "version": "17.11.0",
                    "appVersion": "7.0.11",
                    "name": "redis",
                    "urls": [
                        "harbor.example.com/charts/redis:17.11.0"
                    ],
                    "digest": "sha256:45925becfe9aa2c6c4741c9fe1dd0ddca627894b696755c73830e4ae6b390c35",
                    "releasedAt": "2023-06-09T11:50:48.144176763Z"
                }
            ]
        }
    },
    "apiVersion": "v1"
}
`

const chartsIndexSingleJSON = `
{
    "entries": {
        "common": {
            "versions": [
                {
                    "version": "2.4.0",
                    "appVersion": "2.4.0",
                    "name": "common",
                    "urls": [
                        "harbor.example.com/charts/common:2.4.0"
                    ],
                    "digest": "sha256:c85139bbe83ec5af6201fe1bec39fc0d0db475de41bc74cd729acc5af8eed6ba",
                    "releasedAt": "2023-06-08T12:15:48.149853788Z"
                }
            ]
        }
    },
    "apiVersion": "v1"
}
`

func iconBytes() []byte {
	var b bytes.Buffer
	img := imaging.New(1, 1, color.White)
	err := imaging.Encode(&b, img, imaging.PNG)
	if err != nil {
		return nil
	}
	return b.Bytes()
}

var testChartReadme = "# readme for chart\n\nBest chart in town"
var testChartValues = "image: test"
var testChartSchema = `{"properties": {}}`

func newFakeServer(t *testing.T, responses map[string]*http.Response) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, response := range responses {
			if path == r.URL.Path {
				if response.Header != nil {
					for k, vs := range response.Header {
						for _, v := range vs {
							w.Header().Set(k, v)
						}
					}
				}
				w.WriteHeader(response.StatusCode)
				body := []byte{}
				if response.Body != nil {
					var err error
					body, err = io.ReadAll(response.Body)
					if err != nil {
						t.Fatalf("%+v", err)
					}
				}
				_, err := w.Write(body)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				return
			}
		}
		w.WriteHeader(404)
	}))
}

func Test_syncURLInvalidity(t *testing.T) {
	tests := []struct {
		name    string
		repoURL string
	}{
		{"invalid URL", "not-a-url"},
		{"invalid URL", "https//google.com"},
	}

	fakeServer := newFakeServer(t, nil)
	defer fakeServer.Close()
	pgManager, _, cleanup := getMockManager(t)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getHelmRepo("namespace", "test", tt.repoURL, "", nil, fakeServer.Client(), "my-user-agent", pgManager)
			assert.Error(t, err, tt.name)
		})
	}
}

func Test_getOCIRepo(t *testing.T) {
	grpcClient, f, err := ocicatalog_client.NewClient("test")
	assert.NoError(t, err)
	defer f()
	t.Run("it should add the auth header to the resolver", func(t *testing.T) {
		repo, err := getOCIRepo("namespace", "test", "https://test", "Basic auth", nil, []string{}, &http.Client{}, &grpcClient, nil)
		assert.NoError(t, err)
		helmtest.CheckHeader(t, repo.(*OCIRegistry).puller, "Authorization", "Basic auth")
	})

	t.Run("it should use https for distribution spec API calls if protocol is oci", func(t *testing.T) {
		repo, err := getOCIRepo("namespace", "test", "oci://test", "Basic auth", nil, []string{}, &http.Client{}, &grpcClient, nil)
		assert.NoError(t, err)

		client := repo.(*OCIRegistry).ociCli
		if got, want := client.(*OciAPIClient).RegistryNamespaceUrl.String(), "https://test"; got != want {
			t.Errorf("got: %q, want: %q", got, want)
		}
	})
}

func Test_parseFilters(t *testing.T) {
	t.Run("return rules spec", func(t *testing.T) {
		filters, err := parseFilters(`{"jq":".name == $var1","variables":{"$var1":"wordpress"}}`)
		assert.NoError(t, err)
		assert.Equal(t, filters, &apprepov1alpha1.FilterRuleSpec{
			JQ: ".name == $var1", Variables: map[string]string{"$var1": "wordpress"},
		}, "filters")
	})
}

func Test_fetchRepoIndex(t *testing.T) {
	fakeServer := newFakeServer(t, map[string]*http.Response{
		"/index.yaml":         {StatusCode: 200},
		"/subpath/index.yaml": {StatusCode: 200},
	})
	defer fakeServer.Close()
	addr := fakeServer.URL

	tests := []struct {
		name      string
		url       string
		userAgent string
	}{
		{"valid HTTP URL", addr, "my-user-agent"},
		{"valid trailing URL", addr + "/", "my-user-agent"},
		{"valid subpath URL", addr + "/subpath/", "my-user-agent"},
		{"valid URL with trailing spaces", addr + "/subpath/  ", "my-user-agent"},
		{"valid URL with leading spaces", "  " + addr + "/subpath/", "my-user-agent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netClient := fakeServer.Client()
			_, err := fetchRepoIndex(tt.url, "", netClient, tt.userAgent)
			assert.NoError(t, err)
		})
	}

	validAuthHeader := "Bearer ThisSecretAccessTokenAuthenticatesTheClient"
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == validAuthHeader {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(401)
	}))

	t.Run("authenticated request", func(t *testing.T) {
		netClient := authServer.Client()
		_, err := fetchRepoIndex(authServer.URL, validAuthHeader, netClient, "my-user-agent")
		assert.NoError(t, err)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		_, err := fetchRepoIndex(authServer.URL, "Bearer: not-valid", authServer.Client(), "my-user-agent")
		assert.Error(t, err, errors.New("failed?"))
	})
}

func Test_fetchRepoIndexUserAgent(t *testing.T) {
	tests := []struct {
		name              string
		version           string
		userAgentComment  string
		expectedUserAgent string
	}{
		{"default user agent", "", "", "asset-syncer/devel"},
		{"custom version no app", "1.0", "", "asset-syncer/1.0"},
		{"custom version and app", "1.0", "foo/1.2", "asset-syncer/1.0 (foo/1.2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override global variables used to generate the userAgent

			userAgent := GetUserAgent(tt.version, tt.userAgentComment)

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, req.Header.Get("User-Agent"), tt.expectedUserAgent, "expected user agent")
				_, err := rw.Write([]byte(validRepoIndexYAML))
				if err != nil {
					log.Fatalf("%+v", err)
				}
			}))
			// Close the server when test finishes
			defer server.Close()

			netClient := server.Client()

			_, err := fetchRepoIndex(server.URL, "", netClient, userAgent)
			assert.NoError(t, err)
		})
	}
}

func Test_chartTarballURL(t *testing.T) {
	r := &models.AppRepositoryInternal{Name: "test", URL: "http://testrepo.com"}
	tests := []struct {
		name   string
		cv     models.ChartVersion
		wanted string
	}{
		{"absolute url", models.ChartVersion{URLs: []string{"http://testrepo.com/wordpress-0.1.0.tgz"}}, "http://testrepo.com/wordpress-0.1.0.tgz"},
		{"relative url", models.ChartVersion{URLs: []string{"wordpress-0.1.0.tgz"}}, "http://testrepo.com/wordpress-0.1.0.tgz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, chartTarballURL(r, tt.cv), tt.wanted, "url")
		})
	}
}

func Test_initNetClient(t *testing.T) {
	// Test env
	otherDir, err := os.MkdirTemp("", "ca-registry")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(otherDir)

	// Create cert
	caCert := `-----BEGIN CERTIFICATE-----
MIIC6jCCAdKgAwIBAgIUKVfzA7lfBgSYP8enCVhlm0ql5YwwDQYJKoZIhvcNAQEL
BQAwDTELMAkGA1UEAxMCQ0EwHhcNMTgxMjEyMTQxNzAwWhcNMjMxMjExMTQxNzAw
WjANMQswCQYDVQQDEwJDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
ALZU3fsAgvoUuLSHr24apslaYyuX84wGoZQmtFtQ+A3DF9KL/2nn3yZ6qJPkH0TF
sbObEQRNi+P6vQ3nI/dSNMX5PzMBP2CB6L7zEXzZQEHtAK0Bzva5CKEBGX7OfIKl
aBvs+dzKVJBdb+Oh0maacMwa4QbcD6ejzF90jUbaO65lpQpcL7KQdppKOGNclRaA
hQTV2VsxrV4hH7K9btaTTxso+8W6p8v6X9vf40Ywx72p+SKnGh+FCrOp1gYLBLwo
4SM0OUQHRvqUlj0XhZk5pW0dMRwHcoz1S2GmE5bj4edr4j+zGzGxa2wRGKvM0OCn
Do84AVszTFPmUf+mCl4pJNECAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1Ud
EwEB/wQFMAMBAf8wHQYDVR0OBBYEFI5l5k+MEhrbOQ29dOW1qJhI0yKaMA0GCSqG
SIb3DQEBCwUAA4IBAQByDebUOKzn6jfmXlW62vm09V+ipqId01wm21G9XMtMEqhc
xtun6YwQeTuGPtdepWG+NXuSsiX/HNAHeaumJaaljHhdKDisnMQ0CTnNsu8NPkAl
9iMEB3iXLWkb7+HgfPJAHZVGcMqMxNEMZYHB1Fh0G2Ne376X94+GYJ08qR2C8rUP
BShhMSktB578h4GtPIWSjPhDUWg1fGe7sewR+GPyuL9859hOD0wGm9tUixBKloCu
b90fhqZZ3FqZD7W1qJGKvz/8geqi0noip+uq/dokK1jarRkOVEJP+EvXkHo0tIuc
h251U/Daz6NiQBM9AxyAw6EHm8XAZBvCuebfzyrT
-----END CERTIFICATE-----`
	otherCA := path.Join(otherDir, "ca.crt")
	err = os.WriteFile(otherCA, []byte(caCert), 0644)
	if err != nil {
		t.Error(err)
	}

	_, err = httpclient.NewWithCertFile(otherCA, false)
	if err != nil {
		t.Error(err)
	}
}

func Test_getSha256(t *testing.T) {
	sha, err := getSha256([]byte("this is a test"))
	assert.Equal(t, err, nil, "Unable to get sha")
	assert.Equal(t, sha, "2e99758548972a8e8822ad47fa1017ff72f06f3ff6a016851f45c398732bc50c", "Unable to get sha")
}

func Test_newManager(t *testing.T) {
	tests := []struct {
		name            string
		dbName          string
		dbURL           string
		dbUser          string
		dbPass          string
		expectedManager string
	}{
		{"postgresql database", "assets", "example.com:44124", "postgres", "root", "&{host=example.com port=44124 user=postgres password=root dbname=assets sslmode=disable <nil>}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := dbutils.Config{URL: tt.dbURL, Database: tt.dbName, Username: tt.dbUser, Password: tt.dbPass}
			_, err := newManager(config, "kubeapps")
			assert.NoError(t, err)
		})
	}

}

func Test_fetchAndImportIcon(t *testing.T) {
	repo := &models.AppRepositoryInternal{Name: "test", Namespace: "repo-namespace"}
	validBearer := "Bearer ThisSecretAccessTokenAuthenticatesTheClient"
	repoWithAuthorization := &models.AppRepositoryInternal{Name: "test", Namespace: "repo-namespace", AuthorizationHeader: validBearer, URL: "https://github.com/"}

	svgHeader := http.Header{}
	svgHeader.Set("Content-Type", "image/svg")

	server := newFakeServer(t, map[string]*http.Response{
		"/valid_icon.png": {
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(iconBytes())),
		},
		"/valid_svg_icon.svg": {
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte("<svg width='100' height='100'></svg>"))),
			Header:     svgHeader,
		},
		"/download_fail.png": {
			StatusCode: 500,
		},
		"/invalid_icon.png": {
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte("not a valid png"))),
		},
	})
	defer server.Close()
	t.Run("no icon", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		c := models.Chart{ID: "test/acs-engine-autoscaler"}
		fImporter := fileImporter{pgManager, server.Client()}
		assert.NoError(t, fImporter.fetchAndImportIcon(c, repo, "my-user-agent", false))
	})

	charts, _ := helm.ChartsFromIndex([]byte(validRepoIndexYAML), &models.AppRepository{Name: "test", Namespace: "repo-namespace", URL: server.URL}, false)
	chart := charts[0]

	t.Run("failed download", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		fImporter := fileImporter{pgManager, server.Client()}
		chart.Icon = server.URL + "/download_fail.png"

		assert.Equal(t, fmt.Errorf("GET request to [%s] failed due to status [500]", chart.Icon), fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	t.Run("bad icon", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		fImporter := fileImporter{pgManager, server.Client()}
		chart.Icon = server.URL + "/invalid_icon.png"
		assert.Equal(t, image.ErrFormat, fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	t.Run("valid icon", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = server.URL + "/valid_icon.png"

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, server.Client()}
		assert.NoError(t, fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	t.Run("valid SVG icon", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = server.URL + "/valid_svg_icon.svg"

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, server.Client()}
		assert.NoError(t, fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != validBearer {
			w.WriteHeader(401)
			return
		}
		if strings.HasSuffix(r.RequestURI, "/valid_icon.png") {
			w.WriteHeader(200)
			_, err := w.Write(iconBytes())
			if err != nil {
				t.Fatalf("%+v", err)
			}
			return
		}
		if strings.HasSuffix(r.RequestURI, "/invalid_icon.png") {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(404)
	}))
	defer authServer.Close()
	t.Run("valid icon (not passing through the auth header by default)", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = authServer.URL + "/valid_icon.png"

		fImporter := fileImporter{pgManager, authServer.Client()}
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [401]", charts[0].Icon), fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	t.Run("valid icon (not passing through the auth header)", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = authServer.URL + "/valid_icon.png"

		fImporter := fileImporter{pgManager, authServer.Client()}
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [401]", charts[0].Icon), fImporter.fetchAndImportIcon(chart, repo, "my-user-agent", false))
	})

	t.Run("valid icon (not passing through the auth header if not same domain)", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = authServer.URL + "/valid_icon.png"

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, authServer.Client()}
		// Even though the repo has the auth token, we don't use it to download
		// the icon if the icon is on a different domain.
		repoWithAuthorization.URL = "https://github.com"
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [401]", charts[0].Icon), fImporter.fetchAndImportIcon(chart, repoWithAuthorization, "my-user-agent", false))
	})

	t.Run("valid icon (passing through the auth header if same domain)", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = authServer.URL + "/valid_icon.png"

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, authServer.Client()}
		// If the repo URL matches the domain of the icon, then it's
		// safe to send the creds.
		repoWithAuthorization.URL = authServer.URL
		assert.NoError(t, fImporter.fetchAndImportIcon(chart, repoWithAuthorization, "my-user-agent", false))
	})

	t.Run("valid icon (passing through the auth header)", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		chart.Icon = authServer.URL + "/valid_icon.png"

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, authServer.Client()}
		assert.NoError(t, fImporter.fetchAndImportIcon(chart, repoWithAuthorization, "my-user-agent", true))
	})
}

type fakeRepo struct {
	*models.AppRepositoryInternal
	charts     []models.Chart
	chartFiles models.ChartFiles
}

func (r *fakeRepo) Checksum(ctx context.Context) (string, error) {
	return "checksum", nil
}

func (r *fakeRepo) AppRepository() *models.AppRepositoryInternal {
	return r.AppRepositoryInternal
}

func (r *fakeRepo) SortVersions() {
	// no-op
}

func (r *fakeRepo) Filters() *apprepov1alpha1.FilterRuleSpec {
	return nil
}

func (r *fakeRepo) Charts(ctx context.Context, shallow bool, chartResults chan pullChartResult) ([]string, error) {
	for _, chart := range r.charts {
		chartResults <- pullChartResult{
			Chart: chart,
		}
	}
	close(chartResults)
	return nil, nil
}

func (r *fakeRepo) FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error) {
	return map[string]string{
		models.DefaultValuesKey: r.chartFiles.DefaultValues,
		models.ReadmeKey:        r.chartFiles.Readme,
		models.SchemaKey:        r.chartFiles.Schema,
	}, nil
}

func Test_fetchAndImportFiles(t *testing.T) {
	validAuthHeader := "Bearer ThisSecretAccessTokenAuthenticatesTheClient"

	// Update the URL for the chart version file so that it uses the test server.
	internalRepo := &models.AppRepositoryInternal{Name: "test", Namespace: "repo-namespace", AuthorizationHeader: validAuthHeader}
	charts, _ := helm.ChartsFromIndex([]byte(validRepoIndexYAML), &models.AppRepository{Name: internalRepo.Name, Namespace: internalRepo.Namespace, URL: internalRepo.URL}, false)
	chartVersion := charts[0].ChartVersions[0]
	chartVersionURL, err := url.Parse(chartVersion.URLs[0])
	if err != nil {
		t.Fatalf("%+v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != validAuthHeader {
			w.WriteHeader(401)
			return
		}
		if strings.HasSuffix(r.RequestURI, ".tgz") {
			gzw := gzip.NewWriter(w)
			files := []tartest.TarballFile{{Name: charts[0].Name + "/Chart.yaml", Body: "should be a Chart.yaml here..."}}
			files = append(files, tartest.TarballFile{Name: charts[0].Name + "/values.yaml", Body: testChartValues})
			files = append(files, tartest.TarballFile{Name: charts[0].Name + "/README.md", Body: testChartReadme})
			files = append(files, tartest.TarballFile{Name: charts[0].Name + "/values.schema.json", Body: testChartSchema})
			tartest.CreateTestTarball(gzw, files)
			gzw.Flush()
			return
		}
		w.WriteHeader(200)
		_, err := w.Write([]byte("Foo"))
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}))
	defer server.Close()

	// Ensure that the server URL is used in tests.
	internalRepo.URL = server.URL
	chartVersion.URLs[0] = server.URL + chartVersionURL.Path
	charts[0].Repo.URL = server.URL

	chartID := fmt.Sprintf("%s/%s", charts[0].Repo.Name, charts[0].Name)
	chartFilesID := fmt.Sprintf("%s-%s", chartID, chartVersion.Version)
	chartFiles := models.ChartFiles{
		ID:                      chartFilesID,
		Readme:                  testChartReadme,
		DefaultValues:           testChartValues,
		AdditionalDefaultValues: map[string]string{},
		Schema:                  testChartSchema,
		Repo:                    charts[0].Repo,
		Digest:                  chartVersion.Digest,
	}
	fRepo := &fakeRepo{
		AppRepositoryInternal: internalRepo,
		charts:                charts,
		chartFiles:            chartFiles,
	}

	t.Run("http error", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		mock.ExpectQuery("SELECT EXISTS*").
			WithArgs(chartFilesID, internalRepo.Name, internalRepo.Namespace).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))
		fImporter := fileImporter{pgManager, server.Client()}
		helmRepo := &HelmRepo{
			content:               []byte{},
			AppRepositoryInternal: internalRepo,
			netClient:             server.Client(),
		}
		assert.Error(t, fmt.Errorf("GET request to [https://kubernetes-charts.storage.googleapis.com/acs-engine-autoscaler-2.1.1.tgz] failed due to status [500]"), fImporter.fetchAndImportFiles(charts[0].Name, helmRepo, chartVersion, "my-user-agent", false))
	})

	t.Run("file not found", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		// file does not exist (no rows returned) so insertion goes ahead.
		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, internalRepo.Name, internalRepo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}))
		mock.ExpectQuery("INSERT INTO files *").
			WithArgs(chartID, internalRepo.Name, internalRepo.Namespace, chartFilesID, chartFiles).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

		netClient := server.Client()

		fImporter := fileImporter{pgManager, netClient}
		helmRepo := &HelmRepo{
			content:               []byte{},
			AppRepositoryInternal: internalRepo,
			netClient:             server.Client(),
		}
		err := fImporter.fetchAndImportFiles(chartID, helmRepo, chartVersion, "my-user-agent", false)
		assert.NoError(t, err)
	})

	t.Run("file exists", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, internalRepo.Name, internalRepo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(`true`))

		fImporter := fileImporter{pgManager, server.Client()}
		err := fImporter.fetchAndImportFiles(chartID, fRepo, chartVersion, "my-user-agent", false)
		assert.NoError(t, err)
	})
}

func Test_ociAPICli(t *testing.T) {
	t.Run("TagList - failed request", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/apache/tags/list": {
				StatusCode: 500,
			},
		})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}
		_, err = apiCli.TagList("apache", "my-user-agent")
		if err == nil {
			t.Fatalf("got: nil, want: error")
		}
		errResponse, ok := err.(*errcode.ErrorResponse)
		if !ok {
			t.Fatalf("got: %+v, want: *errcode.ErrorResponse", err)
		}
		if got, want := errResponse.StatusCode, http.StatusInternalServerError; got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})

	t.Run("TagList - successful request", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/apache/tags/list": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"name":"apache","tags":["7.5.1","8.1.1"]}`)),
			},
		})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}
		result, err := apiCli.TagList("apache", "my-user-agent")
		assert.NoError(t, err)
		expectedTagList := &TagList{Name: "apache", Tags: []string{"7.5.1", "8.1.1"}}
		if !cmp.Equal(result, expectedTagList) {
			t.Errorf("Unexpected result %v", cmp.Diff(result, expectedTagList))
		}
	})

	t.Run("IsHelmChart - failed request", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/apache-bad/manifests/7.5.1": {
				StatusCode: 500,
			},
		})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}
		_, err = apiCli.IsHelmChart("apache-bad", "7.5.1", "my-user-agent")

		if err == nil {
			t.Fatalf("got: nil, want: error")
		}
		errResponse, ok := err.(*errcode.ErrorResponse)
		if !ok {
			t.Fatalf("got: %+v, want: *errcode.ErrorResponse", err)
		}
		if got, want := errResponse.StatusCode, http.StatusInternalServerError; got != want {
			t.Errorf("got: %d, want: %d", got, want)
		}
	})

	t.Run("IsHelmChart - successful request", func(t *testing.T) {
		manifest751 := `{"schemaVersion":2,"config":{"mediaType":"other","digest":"sha256:123","size":665}}`
		sha751, err := getSha256([]byte(manifest751))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		sha751 = "sha256:" + sha751
		header751 := http.Header{}
		header751.Set("Docker-Content-Digest", sha751)
		header751.Set("Content-Type", "foo")
		header751.Set("Content-Length", fmt.Sprintf("%d", len(manifest751)))

		manifest811 := `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:456","size":665}}`
		sha811, err := getSha256([]byte(manifest811))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		sha811 = "sha256:" + sha811
		header811 := http.Header{}
		header811.Set("Docker-Content-Digest", sha811)
		header811.Set("Content-Type", "foo")
		header811.Set("Content-Length", fmt.Sprintf("%d", len(manifest811)))
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/test/apache/manifests/7.5.1": {
				StatusCode: 200,
				Header:     header751,
			},
			"/v2/test/apache/blobs/" + sha751: {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(manifest751)),
				Header:     header751,
			},
			"/v2/test/apache/manifests/8.1.1": {
				StatusCode: 200,
				Header:     header811,
			},
			"/v2/test/apache/blobs/" + sha811: {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(manifest811)),
				Header:     header811,
			},
		})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}
		is751, err := apiCli.IsHelmChart("test/apache", "7.5.1", "my-user-agent")
		assert.NoError(t, err)
		if is751 {
			t.Errorf("Tag 7.5.1 should not be a helm chart")
		}
		is811, err := apiCli.IsHelmChart("test/apache", "8.1.1", "my-user-agent")
		assert.NoError(t, err)
		if !is811 {
			t.Errorf("Tag 8.1.1 should be a helm chart")
		}
	})

	t.Run("CatalogAvailable - successful request", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/test/project/charts-index/manifests/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(chartsIndexManifestJSON)),
			},
		})
		defer server.Close()

		urlWithNamespace, err := parseRepoURL(server.URL + "/test/project")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: urlWithNamespace,
			HttpClient:           server.Client(),
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, true; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("CatalogAvailable - returns false for incorrect media type", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/test/project/charts-index/manifests/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"config": {"mediaType": "something-else"}}`)),
			},
		})
		defer server.Close()
		urlWithNamespaceBadMediaType, err := parseRepoURL(server.URL + "/test/project")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: urlWithNamespaceBadMediaType,
			HttpClient:           server.Client(),
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, false; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("CatalogAvailable - returns false for a 404", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/test/project/charts-index/manifests/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"config": {"mediaType": "something-else"}}`)),
			},
		})
		defer server.Close()
		urlWithNamespaceNonExistentBlob, err := parseRepoURL(server.URL + "/test/project")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: urlWithNamespaceNonExistentBlob,
			HttpClient:           server.Client(),
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, false; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("CatalogAvailable - returns true if oci-catalog responds", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		grpcAddr, grpcDouble, closer := ocicatalog_clienttest.SetupTestDouble(t)
		defer closer()
		grpcDouble.Repositories = []*ocicatalog.Repository{
			{
				Name: "apache",
			},
		}
		grpcClient, closer, err := ocicatalog_client.NewClient(grpcAddr)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer closer()

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
			GrpcClient:           grpcClient,
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, true; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("CatalogAvailable - returns false if oci-catalog responds with zero repos", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		grpcAddr, grpcDouble, closer := ocicatalog_clienttest.SetupTestDouble(t)
		defer closer()
		grpcDouble.Repositories = []*ocicatalog.Repository{}
		grpcClient, closer, err := ocicatalog_client.NewClient(grpcAddr)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer closer()

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
			GrpcClient:           grpcClient,
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, false; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("CatalogAvailable - returns false on any other", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}

		got, err := apiCli.CatalogAvailable(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if got, want := got, false; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})

	t.Run("Catalog - successful request", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{
			"/v2/test/project/charts-index/manifests/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(chartsIndexManifestJSON)),
			},
			"/v2/test/project/charts-index/blobs/sha256:f9f7df0ae3f50aaf9ff390034cec4286d2aa43f061ce4bc7aa3c9ac862800aba": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(chartsIndexMultipleJSON)),
			},
		})
		defer server.Close()
		url, err := parseRepoURL(server.URL + "/test/project")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
		}

		got, err := apiCli.Catalog(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := got, []string{"common", "redis"}; !cmp.Equal(got, want) {
			t.Errorf("got: %s, want: %s", got, want)
		}
	})

	t.Run("Catalog - successful request via oci-catalog", func(t *testing.T) {
		server := newFakeServer(t, map[string]*http.Response{})
		defer server.Close()
		url, err := parseRepoURL(server.URL)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		grpcAddr, grpcDouble, closer := ocicatalog_clienttest.SetupTestDouble(t)
		defer closer()
		grpcDouble.Repositories = []*ocicatalog.Repository{
			{
				Name: "common",
			},
			{
				Name: "redis",
			},
		}
		grpcClient, closer, err := ocicatalog_client.NewClient(grpcAddr)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer closer()
		apiCli := &OciAPIClient{
			RegistryNamespaceUrl: url,
			HttpClient:           server.Client(),
			GrpcClient:           grpcClient,
		}

		got, err := apiCli.Catalog(context.Background(), "my-user-agent")
		if err != nil {
			t.Fatal(err)
		}

		if got, want := got, []string{"common", "redis"}; !cmp.Equal(got, want) {
			t.Errorf("got: %s, want: %s", got, want)
		}
	})
}

func Test_OCIRegistry(t *testing.T) {
	chartYAML := `
annotations:
  category: Infrastructure
apiVersion: v2
appVersion: 2.0.0
description: chart description
home: https://kubeapps.com
icon: https://logo.png
keywords:
  - helm
maintainers:
  - email: containers@bitnami.com
    name: Bitnami
name: kubeapps
sources:
  - https://github.com/vmware-tanzu/kubeapps
version: 1.0.0
`
	tests := []struct {
		description      string
		chartName        string
		ociArtifactFiles []tartest.TarballFile
		tags             []string
		expected         []models.Chart
		shallow          bool
	}{
		{
			"Retrieve chart metadata",
			"kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/kubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/test"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							Digest:                  "123",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							AdditionalDefaultValues: map[string]string{},
						},
					},
				},
			},
			false,
		},
		{
			"Retrieve standard files",
			"kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/kubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/test"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:                  "123",
							Readme:                  "chart readme",
							DefaultValues:           "chart values",
							AdditionalDefaultValues: map[string]string{},
							Schema:                  "chart schema",
						},
					},
				},
			},
			false,
		},
		{
			"Retrieve additional values files",
			"kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values-production.yaml", Body: "chart prod values"},
				{Name: "values-staging.yaml", Body: "chart staging values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/kubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/test"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:       "1.0.0",
							AppVersion:    "2.0.0",
							URLs:          []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:        "123",
							Readme:        "chart readme",
							DefaultValues: "chart values",
							Schema:        "chart schema",
							AdditionalDefaultValues: map[string]string{
								"values-production": "chart prod values",
								"values-staging":    "chart staging values",
							},
						},
					},
				},
			},
			false,
		},
		{
			"Retrieve additional values files with more hyphens",
			"kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values-scenario-a.yaml", Body: "scenario a values"},
				{Name: "values-scenario-b.yaml", Body: "scenario b values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/kubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/test"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:       "1.0.0",
							AppVersion:    "2.0.0",
							URLs:          []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:        "123",
							Readme:        "chart readme",
							DefaultValues: "chart values",
							Schema:        "chart schema",
							AdditionalDefaultValues: map[string]string{
								"values-scenario-a": "scenario a values",
								"values-scenario-b": "scenario b values",
							},
						},
					},
				},
			},
			false,
		},
		{
			"A chart with a /",
			"repo/kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:                  "123",
							Readme:                  "chart readme",
							DefaultValues:           "chart values",
							AdditionalDefaultValues: map[string]string{},
							Schema:                  "chart schema",
						},
					},
				},
			},
			false,
		},
		{
			"Multiple chart versions",
			"repo/kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0", "1.1.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:                  "123",
							Readme:                  "chart readme",
							DefaultValues:           "chart values",
							AdditionalDefaultValues: map[string]string{},
							Schema:                  "chart schema",
						},
						{
							// The test passes the one yaml file for both tags,
							// hence the same version number here.
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:                  "123",
							Readme:                  "chart readme",
							DefaultValues:           "chart values",
							AdditionalDefaultValues: map[string]string{},
							Schema:                  "chart schema",
						},
					},
				},
			},
			false,
		},
		{
			"Single chart version for a shallow run",
			"repo/kubeapps",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: chartYAML},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.1.0", "1.0.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "kubeapps",
					Repo:        &models.AppRepository{Name: "test", URL: "http://oci-test/"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:                 "1.0.0",
							AppVersion:              "2.0.0",
							URLs:                    []string{"https://github.com/vmware-tanzu/kubeapps"},
							Digest:                  "123",
							Readme:                  "chart readme",
							DefaultValues:           "chart values",
							AdditionalDefaultValues: map[string]string{},
							Schema:                  "chart schema",
						},
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			w := map[string]*httptest.ResponseRecorder{}
			content := map[string]*bytes.Buffer{}
			for _, tag := range tt.tags {
				recorder := httptest.NewRecorder()
				gzw := gzip.NewWriter(recorder)
				tartest.CreateTestTarball(gzw, tt.ociArtifactFiles)
				gzw.Flush()
				w[tag] = recorder
				content[tag] = recorder.Body
			}

			tags := map[string]*http.Response{}
			for _, tag := range tt.tags {
				tags[fmt.Sprintf("/v2/%s/manifests/%s", tt.chartName, tag)] = &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:123","size":665}}`)),
				}
			}
			tagList, err := json.Marshal(TagList{
				Name: fmt.Sprintf("test/%s", tt.chartName),
				Tags: tt.tags,
			})
			if err != nil {
				t.Fatalf("%+v", err)
			}
			tags[fmt.Sprintf("/v2/%s/tags/list", tt.chartName)] = &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(tagList)),
			}
			server := newFakeServer(t, tags)
			defer server.Close()
			url, err := parseRepoURL(server.URL)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			pgManager, mockDB, cleanup := getMockManager(t)
			defer cleanup()
			mockDB.ExpectQuery("SELECT info FROM charts *").
				WillReturnRows(sqlmock.NewRows([]string{"info"}).
					AddRow(string("{}")))
			chartsRepo := OCIRegistry{
				repositories:          []string{tt.chartName},
				AppRepositoryInternal: &models.AppRepositoryInternal{Name: tt.expected[0].Repo.Name, URL: tt.expected[0].Repo.URL},
				puller: &helmfake.OCIPuller{
					Content:  content,
					Checksum: "123",
				},
				ociCli: &OciAPIClient{
					RegistryNamespaceUrl: url,
					HttpClient:           server.Client(),
				},
				manager: pgManager,
			}
			chartResults := make(chan pullChartResult, 2)
			_, err = chartsRepo.Charts(context.Background(), tt.shallow, chartResults)
			assert.NoError(t, err)

			charts := []models.Chart{}
			for chartsResult := range chartResults {
				charts = append(charts, chartsResult.Chart)
			}
			if !cmp.Equal(charts, tt.expected) {
				t.Errorf("Unexpected result %v", cmp.Diff(tt.expected, charts))
			}
		})
	}

	t.Run("it fetches repositories when not present", func(t *testing.T) {
		content := map[string]*bytes.Buffer{}
		files := []tartest.TarballFile{
			{Name: "Chart.yaml", Body: chartYAML},
			{Name: "README.md", Body: "chart readme"},
			{Name: "values.yaml", Body: "chart values"},
			{Name: "values.schema.json", Body: "chart schema"},
		}
		tag := "1.1.0"
		recorder := httptest.NewRecorder()
		gzw := gzip.NewWriter(recorder)
		tartest.CreateTestTarball(gzw, files)
		gzw.Flush()
		content[tag] = recorder.Body

		tagList, err := json.Marshal(TagList{
			Name: "test/common",
			Tags: []string{"1.1.0"},
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		fakeURIs := map[string]*http.Response{
			"/v2/my-project/common/manifests/1.1.0": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:123","size":665}}`)),
			},
			"/v2/my-project/charts-index/manifests/latest": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(chartsIndexManifestJSON)),
			},
			"/v2/my-project/charts-index/blobs/sha256:f9f7df0ae3f50aaf9ff390034cec4286d2aa43f061ce4bc7aa3c9ac862800aba": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(chartsIndexSingleJSON)),
			},
			"/v2/my-project/common/tags/list": {
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(tagList)),
			},
		}
		server := newFakeServer(t, fakeURIs)
		defer server.Close()
		url, err := parseRepoURL(server.URL + "/my-project")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		pgManager, mockDB, cleanup := getMockManager(t)
		defer cleanup()
		mockDB.ExpectQuery("SELECT info FROM charts *").
			WillReturnRows(sqlmock.NewRows([]string{"info"}).
				AddRow(string("{}")))

		chartsRepo := OCIRegistry{
			repositories:          []string{},
			AppRepositoryInternal: &models.AppRepositoryInternal{Name: "common", URL: server.URL},
			puller: &helmfake.OCIPuller{
				Content:  content,
				Checksum: "123",
			},
			ociCli: &OciAPIClient{
				RegistryNamespaceUrl: url,
				HttpClient:           server.Client(),
			},
			manager: pgManager,
		}
		chartResults := make(chan pullChartResult, 2)
		_, err = chartsRepo.Charts(context.Background(), true, chartResults)
		assert.NoError(t, err)

		charts := []models.Chart{}
		for chartResult := range chartResults {
			charts = append(charts, chartResult.Chart)
		}
		if len(charts) != 1 && charts[0].Name != "common" {
			t.Errorf("got: %+v", charts)
		}
	})

	t.Run("FetchFiles - It returns the stored files", func(t *testing.T) {
		files := map[string]string{
			models.DefaultValuesKey: "values text",
			models.ReadmeKey:        "readme text",
			models.SchemaKey:        "schema text",
		}
		repo := OCIRegistry{}
		result, err := repo.FetchFiles(models.ChartVersion{
			DefaultValues: files[models.DefaultValuesKey],
			Readme:        files[models.ReadmeKey],
			Schema:        files[models.SchemaKey],
		}, "my-user-agent", false)
		assert.NoError(t, err)
		assert.Equal(t, result, files, "expected files")
	})
}

func Test_filterMatches(t *testing.T) {
	tests := []struct {
		description string
		input       models.Chart
		rule        apprepov1alpha1.FilterRuleSpec
		expected    bool
		expectedErr error
	}{
		{
			"should match a named chart",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1", Variables: map[string]string{"$var1": "foo"},
			},
			true,
			nil,
		},
		{
			"should not match a named chart",
			models.Chart{
				Name: "bar",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1", Variables: map[string]string{"$var1": "foo"},
			},
			false,
			nil,
		},
		{
			"an invalid rule cause to return an empty set",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: "not a rule",
			},
			false,
			fmt.Errorf(`unable to parse jq query: unexpected token "a"`),
		},
		{
			"an invalid number of vars cause to return an empty set",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1",
			},
			false,
			fmt.Errorf(`unable to compile jq: variable not defined: $var1`),
		},
		{
			"the query doesn't return a boolean",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name`,
			},
			false,
			fmt.Errorf(`unable to convert jq result to boolean. Got: foo`),
		},
		{
			"matches without vars",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name == "foo"`,
			},
			true,
			nil,
		},
		{
			"matches negatively without vars",
			models.Chart{
				Name: "bar",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name == "foo"`,
			},
			false,
			nil,
		},
		{
			"filters a maintainer name",
			models.Chart{
				Name: "foo", Maintainers: []chart.Maintainer{{Name: "Bitnami"}},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".maintainers | any(.name == $var1)", Variables: map[string]string{"$var1": "Bitnami"},
			},
			true,
			nil,
		},
		{
			"filter matches negatively a maintainer name",
			models.Chart{
				Name: "bar", Maintainers: []chart.Maintainer{{Name: "Hackers"}},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".maintainers | any(.name == $var1)", Variables: map[string]string{"$var1": "Bitnami"},
			},
			false,
			nil,
		},
		{
			"excludes a value",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1 | not", Variables: map[string]string{"$var1": "foo"},
			},
			false,
			nil,
		},
		{
			"matches against a regex",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name | test($var1)`, Variables: map[string]string{"$var1": ".*oo.*"},
			},
			true,
			nil,
		},
		{
			"ignores an empty rule",
			models.Chart{
				Name: "foo",
			},
			apprepov1alpha1.FilterRuleSpec{},
			true,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			res, err := filterMatches(tt.input, &tt.rule)
			if err != nil {
				if tt.expectedErr == nil || err.Error() != tt.expectedErr.Error() {
					t.Fatalf("Unexpected error %v", err)
				}
			}
			if !cmp.Equal(res, tt.expected) {
				t.Errorf("Unexpected result: %v", cmp.Diff(res, tt.expected))
			}
		})
	}
}

func TestUnescapeChartsData(t *testing.T) {
	tests := []struct {
		description string
		input       models.Chart
		expected    models.Chart
	}{
		{
			"chart with encoded spaces in id",
			models.Chart{
				ID: "foo%20bar",
			},
			models.Chart{
				ID: "foo bar",
			},
		},
		{
			"chart with encoded spaces in name",
			models.Chart{
				Name: "foo%20bar",
			},
			models.Chart{
				Name: "foo bar",
			},
		},
		{
			"chart with mixed encoding in name",
			models.Chart{
				Name: "test/foo%20bar",
			},
			models.Chart{
				Name: "test/foo bar",
			},
		},
		{
			"chart with no encoding nor spaces",
			models.Chart{
				Name: "test/foobar",
			},
			models.Chart{
				Name: "test/foobar",
			},
		},
		{
			"chart with unencoded spaces",
			models.Chart{
				Name: "test/foo bar",
			},
			models.Chart{
				Name: "test/foo bar",
			},
		},
		{
			"chart with encoded chars in name",
			models.Chart{
				Name: "foo%23bar%2ebar",
			},
			models.Chart{
				Name: "foo#bar.bar",
			},
		},
		{
			"slashes in the chart name are not unescaped",
			models.Chart{
				ID:   "repo-name/project1%2Ffoo%20bar",
				Name: "project1%2Ffoo%20bar",
			},
			models.Chart{
				ID:   "repo-name/project1%2Ffoo bar",
				Name: "project1%2Ffoo bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			res := unescapeChartData(tt.input)
			if !cmp.Equal(res, tt.expected) {
				t.Errorf("Unexpected result: %v", cmp.Diff(res, tt.expected))
			}
		})
	}
}

func TestHelmRepoAppliesUnescape(t *testing.T) {
	repo := &models.AppRepositoryInternal{Name: "test", Namespace: "repo-namespace", URL: "http://testrepo.com"}
	expectedRepo := &models.AppRepository{Name: repo.Name, Namespace: repo.Namespace, URL: repo.URL}
	repoIndexYAMLBytes, _ := os.ReadFile("testdata/helm-index-spaces.yaml")
	repoIndexYAML := string(repoIndexYAMLBytes)
	expectedCharts := []models.Chart{
		{
			ID:            "test/chart$with$chars",
			Name:          "chart$with$chars",
			Repo:          expectedRepo,
			Maintainers:   []chart.Maintainer{},
			ChartVersions: []models.ChartVersion{{AppVersion: "v1"}},
		},
		{
			ID:            "test/chart with spaces",
			Name:          "chart with spaces",
			Repo:          expectedRepo,
			Maintainers:   []chart.Maintainer{},
			ChartVersions: []models.ChartVersion{{AppVersion: "v1"}},
		},
		{
			ID:            "test/chart#with#hashes",
			Name:          "chart#with#hashes",
			Repo:          expectedRepo,
			Maintainers:   []chart.Maintainer{},
			ChartVersions: []models.ChartVersion{{AppVersion: "v3"}},
		},
		{
			ID:            "test/chart-without-spaces",
			Name:          "chart-without-spaces",
			Repo:          expectedRepo,
			Maintainers:   []chart.Maintainer{},
			ChartVersions: []models.ChartVersion{{AppVersion: "v2"}},
		},
	}
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()
	mock.ExpectQuery("SELECT info FROM charts").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow("{}"))
	helmRepo := &HelmRepo{
		content:               []byte(repoIndexYAML),
		AppRepositoryInternal: repo,
		manager:               pgManager,
	}
	t.Run("Helm repo applies unescaping to chart data", func(t *testing.T) {
		chartResults := make(chan pullChartResult, 2)
		_, err := helmRepo.Charts(context.Background(), false, chartResults)
		assert.NoError(t, err)
		charts := []models.Chart{}
		for cr := range chartResults {
			charts = append(charts, cr.Chart)
		}

		if !cmp.Equal(charts, expectedCharts) {
			t.Errorf("Unexpected result: %v", cmp.Diff(expectedCharts, charts))
		}
	})
}

func Test_isURLDomainEqual(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			"it returns false if a url is malformed",
			"abc",
			"https://bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			false,
		},
		{
			"it returns false if they are under different subdomains",
			"https://bitnami.com/bitnami/index.yaml",
			"https://charts.bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			false,
		},
		{
			"it returns false if they are under the same domain but using different schema",
			"http://charts.bitnami.com/bitnami/airflow-10.2.0.tgz",
			"https://charts.bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			false,
		},
		{
			"it returns false if attempting a CRLF injection",
			"https://charts.bitnami.com",
			"https://charts.bitnami.com%0A%0Ddevil.com",
			false,
		},
		{
			"it returns false if attempting a SSRF",
			"https://ⓖⓞⓞⓖⓛⓔ.com",
			"https://google.com",
			false,
		},
		{
			"it returns false if attempting a SSRF",
			"https://wordpress.com",
			"https://wordpreß.com",
			false,
		},
		{
			"it returns false if attempting a SSRF",
			"http://foo@127.0.0.1 @bitnami.com:11211/",
			"https://127.0.0.1:11211",
			false,
		},
		{
			"it returns false if attempting a SSRF",
			"https://foo@evil.com@charts.bitnami.com",
			"https://evil.com",
			false,
			// should be careful, curl would send a request to evil.com
			// but the go net/url parser detects charts.bitnami.com
		},
		{
			"it returns false if attempting a SSRF",
			"https://charts.bitnami.com#@evil.com",
			"https://charts.bitnami.com",
			false,
		},
		{
			"it returns true if they are under the same domain",
			"https://charts.bitnami.com/bitnami/airflow-10.2.0.tgz",
			"https://charts.bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			true,
		},
		{
			"it returns false if they are under the same domain but different ports",
			"https://charts.bitnami.com:8080/bitnami/airflow-10.2.0.tgz",
			"https://charts.bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			false,
		},
		{
			"it returns true if they are under the same domain",
			"https://charts.bitnami.com/bitnami/airflow-10.2.0.tgz",
			"https://charts.bitnami.com/assets/stacks/airflow/img/airflow-stack-220x234.png",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := isURLDomainEqual(tt.url1, tt.url2)
			if !cmp.Equal(res, tt.expected) {
				t.Errorf("Unexpected result: %v", cmp.Diff(res, tt.expected))
			}
		})
	}
}

func TestOrderedChartVersions(t *testing.T) {
	testCases := []struct {
		name          string
		chartVersions []models.ChartVersion
		expected      []models.ChartVersion
	}{
		{
			name: "re-orders an unordered slice",
			chartVersions: []models.ChartVersion{
				{
					Version: "1.2.3",
				},
				{
					Version: "1.2.2",
				},
				{
					Version: "1.2.4",
				},
			},
			expected: []models.ChartVersion{
				{
					Version: "1.2.4",
				},
				{
					Version: "1.2.3",
				},
				{
					Version: "1.2.2",
				},
			},
		},
		{
			name: "an unparsable version is shifted to the end",
			chartVersions: []models.ChartVersion{
				{
					Version: "1.2.4",
				},
				{
					Version: "not-a-version",
				},
				{
					Version: "1.2.2",
				},
			},
			expected: []models.ChartVersion{
				{
					Version: "1.2.4",
				},
				{
					Version: "1.2.2",
				},
				{
					Version: "not-a-version",
				},
			},
		},
		{
			name: "a combination of unorderd versions andan unparsable version is ordered correctly",
			chartVersions: []models.ChartVersion{
				{
					Version: "1.2.2",
				},
				{
					Version: "1.2.4",
				},
				{
					Version: "not-a-version",
				},
				{
					Version: "1.2.3",
				},
			},
			expected: []models.ChartVersion{
				{
					Version: "1.2.4",
				},
				{
					Version: "1.2.3",
				},
				{
					Version: "1.2.2",
				},
				{
					Version: "not-a-version",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chartVersions := tc.chartVersions

			orderedChartVersions(chartVersions)

			if !cmp.Equal(chartVersions, tc.expected) {
				t.Errorf("%s", cmp.Diff(tc.expected, chartVersions))
			}
		})
	}
}
