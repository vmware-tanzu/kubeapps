// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/disintegration/imaging"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	helmfake "github.com/vmware-tanzu/kubeapps/pkg/helm/fake"
	helmtest "github.com/vmware-tanzu/kubeapps/pkg/helm/test"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	tartest "github.com/vmware-tanzu/kubeapps/pkg/tarutil/test"
	"helm.sh/helm/v3/pkg/chart"
	log "k8s.io/klog/v2"
)

var validRepoIndexYAMLBytes, _ = os.ReadFile("testdata/valid-index.yaml")
var validRepoIndexYAML = string(validRepoIndexYAMLBytes)

type badHTTPClient struct {
	errMsg string
}

func (h *badHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	w.WriteHeader(500)
	if len(h.errMsg) > 0 {
		_, err := w.Write([]byte(h.errMsg))
		if err != nil {
			return nil, err
		}
	}
	return w.Result(), nil
}

type goodHTTPClient struct{}

func (h *goodHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	// Don't accept trailing slashes
	if strings.HasPrefix(req.URL.Path, "//") {
		w.WriteHeader(500)
	}

	// Sending an empty Authorization header is not valid for http spec,
	// some servers returning 401 for public resources in this case.
	if v, ok := req.Header["Authorization"]; ok && len(v) == 1 && v[0] == "" {
		w.WriteHeader(401)
	}

	// If subpath repo URL test, check that index.yaml is correctly added to the
	// subpath
	if req.URL.Host == "subpath.test" && req.URL.Path != "/subpath/index.yaml" {
		w.WriteHeader(500)
	}

	_, err := w.Write([]byte(validRepoIndexYAML))
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return w.Result(), nil
}

type goodAuthenticatedHTTPClient struct{}

func (h *goodAuthenticatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()

	// Ensure we're sending an Authorization header
	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(401)
	} else if !strings.Contains(req.Header.Get("Authorization"), "Bearer ThisSecretAccessTokenAuthenticatesTheClient") {
		// Ensure we're sending the right Authorization header
		w.WriteHeader(403)
	} else {
		_, err := w.Write(iconBytes())
		if err != nil {
			log.Fatalf("%+v", err)
		}
	}
	return w.Result(), nil
}

type authenticatedHTTPClient struct{}

func (h *authenticatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()

	// Ensure we're sending the right Authorization header
	if !strings.Contains(req.Header.Get("Authorization"), "Bearer ThisSecretAccessTokenAuthenticatesTheClient") {
		w.WriteHeader(500)
	}
	_, err := w.Write([]byte(validRepoIndexYAML))
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return w.Result(), nil
}

type badIconClient struct{}

func (h *badIconClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	_, err := w.Write([]byte("not-an-image"))
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return w.Result(), nil
}

type goodIconClient struct{}

func iconBytes() []byte {
	var b bytes.Buffer
	img := imaging.New(1, 1, color.White)
	err := imaging.Encode(&b, img, imaging.PNG)
	if err != nil {
		return nil
	}
	return b.Bytes()
}

func (h *goodIconClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	_, err := w.Write(iconBytes())
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return w.Result(), nil
}

type svgIconClient struct{}

func (h *svgIconClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	_, err := w.Write([]byte("<svg width='100' height='100'></svg>"))
	if err != nil {
		log.Fatalf("%+v", err)
	}
	res := w.Result()
	res.Header.Set("Content-Type", "image/svg")
	return res, nil
}

type goodTarballClient struct {
	c          models.Chart
	skipReadme bool
	skipValues bool
	skipSchema bool
}

var testChartReadme = "# readme for chart\n\nBest chart in town"
var testChartValues = "image: test"
var testChartSchema = `{"properties": {}}`

func (h *goodTarballClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	gzw := gzip.NewWriter(w)
	files := []tartest.TarballFile{{Name: h.c.Name + "/Chart.yaml", Body: "should be a Chart.yaml here..."}}
	if !h.skipValues {
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/values.yaml", Body: testChartValues})
	}
	if !h.skipReadme {
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/README.md", Body: testChartReadme})
	}
	if !h.skipSchema {
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/values.schema.json", Body: testChartSchema})
	}
	tartest.CreateTestTarball(gzw, files)
	gzw.Flush()
	return w.Result(), nil
}

type authenticatedTarballClient struct {
	c models.Chart
}

func (h *authenticatedTarballClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()

	// Ensure we're sending the right Authorization header
	if !strings.Contains(req.Header.Get("Authorization"), "Bearer ThisSecretAccessTokenAuthenticatesTheClient") {
		w.WriteHeader(500)
	} else {
		gzw := gzip.NewWriter(w)
		files := []tartest.TarballFile{{Name: h.c.Name + "/Chart.yaml", Body: "should be a Chart.yaml here..."}}
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/values.yaml", Body: testChartValues})
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/README.md", Body: testChartReadme})
		files = append(files, tartest.TarballFile{Name: h.c.Name + "/values.schema.json", Body: testChartSchema})
		tartest.CreateTestTarball(gzw, files)
		gzw.Flush()
	}
	return w.Result(), nil
}

func Test_syncURLInvalidity(t *testing.T) {
	tests := []struct {
		name    string
		repoURL string
	}{
		{"invalid URL", "not-a-url"},
		{"invalid URL", "https//google.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getHelmRepo("namespace", "test", tt.repoURL, "", nil, &goodHTTPClient{}, "my-user-agent")
			assert.Error(t, err, tt.name)
		})
	}
}

func Test_getOCIRepo(t *testing.T) {
	t.Run("it should add the auth header to the resolver", func(t *testing.T) {
		repo, err := getOCIRepo("namespace", "test", "https://test", "Basic auth", nil, []string{}, &http.Client{})
		assert.NoError(t, err)
		helmtest.CheckHeader(t, repo.(*OCIRegistry).puller, "Authorization", "Basic auth")
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
	tests := []struct {
		name      string
		url       string
		userAgent string
	}{
		{"valid HTTP URL", "http://my.examplerepo.com", "my-user-agent"},
		{"valid HTTPS URL", "https://my.examplerepo.com", "my-user-agent"},
		{"valid trailing URL", "https://my.examplerepo.com/", "my-user-agent"},
		{"valid subpath URL", "https://subpath.test/subpath/", "my-user-agent"},
		{"valid URL with trailing spaces", "https://subpath.test/subpath/  ", "my-user-agent"},
		{"valid URL with leading spaces", "  https://subpath.test/subpath/", "my-user-agent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netClient := &goodHTTPClient{}
			_, err := fetchRepoIndex(tt.url, "", netClient, tt.userAgent)
			assert.NoError(t, err)
		})
	}

	t.Run("authenticated request", func(t *testing.T) {
		netClient := &authenticatedHTTPClient{}
		_, err := fetchRepoIndex("https://my.examplerepo.com", "Bearer ThisSecretAccessTokenAuthenticatesTheClient", netClient, "my-user-agent")
		assert.NoError(t, err)
	})

	t.Run("failed request", func(t *testing.T) {
		netClient := &badHTTPClient{}
		_, err := fetchRepoIndex("https://my.examplerepo.com", "", netClient, "my-user-agent")
		assert.Error(t, err, errors.New("failed request"))
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
	r := &models.RepoInternal{Name: "test", URL: "http://testrepo.com"}
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
	repo := &models.RepoInternal{Name: "test", Namespace: "repo-namespace"}
	repoWithAuthorization := &models.RepoInternal{Name: "test", Namespace: "repo-namespace", AuthorizationHeader: "Bearer ThisSecretAccessTokenAuthenticatesTheClient", URL: "https://github.com/"}

	t.Run("no icon", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		c := models.Chart{ID: "test/acs-engine-autoscaler"}
		fImporter := fileImporter{pgManager, &goodHTTPClient{}}
		assert.NoError(t, fImporter.fetchAndImportIcon(c, repo, "my-user-agent", false))
	})

	charts, _ := helm.ChartsFromIndex([]byte(validRepoIndexYAML), &models.Repo{Name: "test", Namespace: "repo-namespace", URL: "http://testrepo.com"}, false)

	t.Run("failed download", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &badHTTPClient{}
		fImporter := fileImporter{pgManager, netClient}
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [500]", charts[0].Icon), fImporter.fetchAndImportIcon(charts[0], repo, "my-user-agent", false))
	})

	t.Run("bad icon", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &badIconClient{}
		c := charts[0]
		fImporter := fileImporter{pgManager, netClient}
		assert.Error(t, image.ErrFormat, fImporter.fetchAndImportIcon(c, repo, "my-user-agent", false))
	})

	t.Run("valid icon", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &goodIconClient{}

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, netClient}
		assert.NoError(t, fImporter.fetchAndImportIcon(charts[0], repo, "my-user-agent", false))
	})

	t.Run("valid SVG icon", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &svgIconClient{}
		c := models.Chart{
			ID:   "foo",
			Icon: "https://foo/bar/logo.svg",
			Repo: &models.Repo{Name: repo.Name, Namespace: repo.Namespace},
		}

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("foo", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, netClient}
		assert.NoError(t, fImporter.fetchAndImportIcon(c, repo, "my-user-agent", false))
	})

	t.Run("valid icon (not passing through the auth header by default)", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &goodAuthenticatedHTTPClient{}

		fImporter := fileImporter{pgManager, netClient}
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [401]", charts[0].Icon), fImporter.fetchAndImportIcon(charts[0], repo, "my-user-agent", false))
	})

	t.Run("valid icon (not passing through the auth header)", func(t *testing.T) {
		pgManager, _, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &goodAuthenticatedHTTPClient{}

		fImporter := fileImporter{pgManager, netClient}
		assert.Error(t, fmt.Errorf("GET request to [%s] failed due to status [401]", charts[0].Icon), fImporter.fetchAndImportIcon(charts[0], repo, "my-user-agent", false))
	})

	t.Run("valid icon (passing through the auth header if same domain)", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &goodAuthenticatedHTTPClient{}

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, netClient}
		assert.NoError(t, fImporter.fetchAndImportIcon(charts[0], repoWithAuthorization, "my-user-agent", false))
	})

	t.Run("valid icon (passing through the auth header)", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		netClient := &goodAuthenticatedHTTPClient{}

		mock.ExpectQuery("UPDATE charts SET info *").
			WithArgs("test/acs-engine-autoscaler", "repo-namespace", "test").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		fImporter := fileImporter{pgManager, netClient}
		assert.NoError(t, fImporter.fetchAndImportIcon(charts[0], repoWithAuthorization, "my-user-agent", true))
	})
}

type fakeRepo struct {
	*models.RepoInternal
	charts     []models.Chart
	chartFiles models.ChartFiles
}

func (r *fakeRepo) Checksum() (string, error) {
	return "checksum", nil
}

func (r *fakeRepo) Repo() *models.RepoInternal {
	return r.RepoInternal
}

func (r *fakeRepo) FilterIndex() {
	// no-op
}

func (r *fakeRepo) Charts(shallow bool) ([]models.Chart, error) {
	return r.charts, nil
}

func (r *fakeRepo) FetchFiles(name string, cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error) {
	return map[string]string{
		models.ValuesKey: r.chartFiles.Values,
		models.ReadmeKey: r.chartFiles.Readme,
		models.SchemaKey: r.chartFiles.Schema,
	}, nil
}

func Test_fetchAndImportFiles(t *testing.T) {
	repo := &models.RepoInternal{Name: "test", Namespace: "repo-namespace", URL: "http://testrepo.com"}
	charts, _ := helm.ChartsFromIndex([]byte(validRepoIndexYAML), &models.Repo{Name: repo.Name, Namespace: repo.Namespace, URL: repo.URL}, false)
	chartVersion := charts[0].ChartVersions[0]
	chartID := fmt.Sprintf("%s/%s", charts[0].Repo.Name, charts[0].Name)
	chartFilesID := fmt.Sprintf("%s-%s", chartID, chartVersion.Version)
	chartFiles := models.ChartFiles{
		ID:     chartFilesID,
		Readme: testChartReadme,
		Values: testChartValues,
		Schema: testChartSchema,
		Repo:   charts[0].Repo,
		Digest: chartVersion.Digest,
	}
	fRepo := &fakeRepo{
		RepoInternal: repo,
		charts:       charts,
		chartFiles:   chartFiles,
	}

	t.Run("http error", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		mock.ExpectQuery("SELECT EXISTS*").
			WithArgs(chartFilesID, repo.Name, repo.Namespace).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))
		netClient := &badHTTPClient{}
		fImporter := fileImporter{pgManager, netClient}
		helmRepo := &HelmRepo{
			content:      []byte{},
			RepoInternal: repo,
			netClient:    netClient,
		}
		assert.Error(t, fmt.Errorf("GET request to [https://kubernetes-charts.storage.googleapis.com/acs-engine-autoscaler-2.1.1.tgz] failed due to status [500]"), fImporter.fetchAndImportFiles(charts[0].Name, helmRepo, chartVersion, "my-user-agent", false))
	})

	t.Run("file not found", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		files := models.ChartFiles{
			ID:     chartFilesID,
			Readme: "",
			Values: "",
			Schema: "",
			Repo:   charts[0].Repo,
			Digest: chartVersion.Digest,
		}

		// file does not exist (no rows returned) so insertion goes ahead.
		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, repo.Name, repo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}))
		mock.ExpectQuery("INSERT INTO files *").
			WithArgs(chartID, repo.Name, repo.Namespace, chartFilesID, files).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

		netClient := &goodTarballClient{c: charts[0], skipValues: true, skipReadme: true, skipSchema: true}

		fImporter := fileImporter{pgManager, netClient}
		helmRepo := &HelmRepo{
			content:      []byte{},
			RepoInternal: repo,
			netClient:    netClient,
		}
		err := fImporter.fetchAndImportFiles(charts[0].Name, helmRepo, chartVersion, "my-user-agent", false)
		assert.NoError(t, err)
	})

	t.Run("authenticated request", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		// file does not exist (no rows returned) so insertion goes ahead.
		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, repo.Name, repo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}))
		mock.ExpectQuery("INSERT INTO files *").
			WithArgs(chartID, repo.Name, repo.Namespace, chartFilesID, chartFiles).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

		netClient := &authenticatedTarballClient{c: charts[0]}

		fImporter := fileImporter{pgManager, netClient}

		r := &models.RepoInternal{Name: repo.Name, Namespace: repo.Namespace, URL: repo.URL, AuthorizationHeader: "Bearer ThisSecretAccessTokenAuthenticatesTheClient"}
		repo := &HelmRepo{
			RepoInternal: r,
			content:      []byte{},
			netClient:    netClient,
		}
		err := fImporter.fetchAndImportFiles(charts[0].Name, repo, chartVersion, "my-user-agent", true)
		assert.NoError(t, err)
	})

	t.Run("valid tarball", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, repo.Name, repo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}))
		mock.ExpectQuery("INSERT INTO files *").
			WithArgs(chartID, repo.Name, repo.Namespace, chartFilesID, chartFiles).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

		netClient := &goodTarballClient{c: charts[0]}
		fImporter := fileImporter{pgManager, netClient}

		err := fImporter.fetchAndImportFiles(charts[0].Name, fRepo, chartVersion, "my-user-agent", false)
		assert.NoError(t, err)
	})

	t.Run("file exists", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()

		mock.ExpectQuery(`SELECT EXISTS*`).
			WithArgs(chartFilesID, repo.Name, repo.Namespace, chartVersion.Digest).
			WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(`true`))

		fImporter := fileImporter{pgManager, &goodHTTPClient{}}
		err := fImporter.fetchAndImportFiles(charts[0].Name, fRepo, chartVersion, "my-user-agent", false)
		assert.NoError(t, err)
	})
}

type goodOCIAPIHTTPClient struct {
	response       string
	responseByPath map[string]string
}

func (h *goodOCIAPIHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	// Don't accept trailing slashes
	if strings.HasPrefix(req.URL.Path, "//") {
		w.WriteHeader(500)
	}

	if r, ok := h.responseByPath[req.URL.Path]; ok {
		_, err := w.Write([]byte(r))
		if err != nil {
			log.Fatalf("%+v", err)
		}
	} else {
		_, err := w.Write([]byte(h.response))
		if err != nil {
			log.Fatalf("%+v", err)
		}
	}
	return w.Result(), nil
}

type authenticatedOCIAPIHTTPClient struct {
	response string
}

func (h *authenticatedOCIAPIHTTPClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()

	// Ensure we're sending the right Authorization header
	if req.Header.Get("Authorization") != "Bearer ThisSecretAccessTokenAuthenticatesTheClient" {
		w.WriteHeader(500)
	}
	_, err := w.Write([]byte(h.response))
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return w.Result(), nil
}

func Test_ociAPICli(t *testing.T) {
	url, _ := parseRepoURL("http://oci-test")

	t.Run("TagList - failed request", func(t *testing.T) {
		apiCli := &ociAPICli{
			url: url,
			netClient: &badHTTPClient{
				errMsg: "forbidden",
			},
		}
		_, err := apiCli.TagList("apache", "my-user-agent")
		assert.Error(t, fmt.Errorf("GET request to [http://oci-test/v2/apache/tags/list] failed due to status [500]: forbidden"), err)
	})

	t.Run("TagList - successful request", func(t *testing.T) {
		apiCli := &ociAPICli{
			url: url,
			netClient: &goodOCIAPIHTTPClient{
				response: `{"name":"test/apache","tags":["7.5.1","8.1.1"]}`,
			},
		}
		result, err := apiCli.TagList("apache", "my-user-agent")
		assert.NoError(t, err)
		expectedTagList := &TagList{Name: "test/apache", Tags: []string{"7.5.1", "8.1.1"}}
		if !cmp.Equal(result, expectedTagList) {
			t.Errorf("Unexpected result %v", cmp.Diff(result, expectedTagList))
		}
	})

	t.Run("TagList with auth - failure", func(t *testing.T) {
		apiCli := &ociAPICli{
			url:        url,
			authHeader: "Bearer wrong",
			netClient:  &authenticatedOCIAPIHTTPClient{},
		}
		_, err := apiCli.TagList("apache", "my-user-agent")
		assert.Error(t, fmt.Errorf("GET request to [http://oci-test/v2/apache/tags/list] failed due to status [500]"), err)
	})

	t.Run("TagList with auth - success", func(t *testing.T) {
		apiCli := &ociAPICli{
			url:        url,
			authHeader: "Bearer ThisSecretAccessTokenAuthenticatesTheClient",
			netClient: &authenticatedOCIAPIHTTPClient{
				response: `{"name":"test/apache","tags":["7.5.1","8.1.1"]}`,
			},
		}
		result, err := apiCli.TagList("apache", "my-user-agent")
		assert.NoError(t, err)
		expectedTagList := &TagList{Name: "test/apache", Tags: []string{"7.5.1", "8.1.1"}}
		if !cmp.Equal(result, expectedTagList) {
			t.Errorf("Unexpected result %v", cmp.Diff(result, expectedTagList))
		}
	})

	t.Run("IsHelmChart - failed request", func(t *testing.T) {
		apiCli := &ociAPICli{
			url:       url,
			netClient: &badHTTPClient{},
		}
		_, err := apiCli.IsHelmChart("apache", "7.5.1", "my-user-agent")
		assert.Error(t, fmt.Errorf("GET request to [http://oci-test/v2/apache/manifests/7.5.1] failed due to status [500]"), err)
	})

	t.Run("IsHelmChart - successful request", func(t *testing.T) {
		apiCli := &ociAPICli{
			url: url,
			netClient: &goodOCIAPIHTTPClient{
				responseByPath: map[string]string{
					// 7.5.1 is not a chart
					"/v2/test/apache/manifests/7.5.1": `{"schemaVersion":2,"config":{"mediaType":"other","digest":"sha256:123","size":665}}`,
					"/v2/test/apache/manifests/8.1.1": `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:123","size":665}}`,
				},
			},
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
}

type fakeOCIAPICli struct {
	tagList *TagList
	err     error
}

func (o *fakeOCIAPICli) TagList(appName, userAgent string) (*TagList, error) {
	return o.tagList, o.err
}

func (o *fakeOCIAPICli) IsHelmChart(appName, tag, userAgent string) (bool, error) {
	return true, o.err
}

func Test_OCIRegistry(t *testing.T) {
	repo := OCIRegistry{
		repositories: []string{"apache", "jenkins"},
		RepoInternal: &models.RepoInternal{
			URL: "http://oci-test",
		},
	}

	t.Run("Checksum - failed request", func(t *testing.T) {
		repo.ociCli = &fakeOCIAPICli{err: fmt.Errorf("request failed")}
		_, err := repo.Checksum()
		assert.Error(t, fmt.Errorf("request failed"), err)
	})

	t.Run("Checksum - success", func(t *testing.T) {
		repo.ociCli = &fakeOCIAPICli{
			tagList: &TagList{Name: "test/apache", Tags: []string{"1.0.0", "1.1.0"}},
		}
		checksum, err := repo.Checksum()
		assert.NoError(t, err)
		assert.Equal(t, checksum, "b1b1ae17ddc8f83606acb8a175025a264e8634bb174b6e6a5799bdb5d20eaa58", "checksum")
	})

	t.Run("Checksum - stores the list of tags", func(t *testing.T) {
		emptyRepo := OCIRegistry{
			repositories: []string{"apache"},
			RepoInternal: &models.RepoInternal{
				URL: "http://oci-test",
			},
			ociCli: &fakeOCIAPICli{
				tagList: &TagList{Name: "test/apache", Tags: []string{"1.0.0", "1.1.0"}},
			},
		}
		_, err := emptyRepo.Checksum()
		assert.NoError(t, err)
		assert.Equal(t, emptyRepo.tags, map[string]TagList{
			"apache": {Name: "test/apache", Tags: []string{"1.0.0", "1.1.0"}},
		}, "expected tags")
	})

	t.Run("FilterIndex - order tags by semver", func(t *testing.T) {
		repo := OCIRegistry{
			repositories: []string{"apache"},
			RepoInternal: &models.RepoInternal{
				URL: "http://oci-test",
			},
			tags: map[string]TagList{
				"apache": {Name: "test/apache", Tags: []string{"1.0.0", "2.0.0", "1.1.0"}},
			},
			ociCli: &fakeOCIAPICli{},
		}
		repo.FilterIndex()
		assert.Equal(t, repo.tags, map[string]TagList{
			"apache": {Name: "test/apache", Tags: []string{"2.0.0", "1.1.0", "1.0.0"}},
		}, "tag list")
	})

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
					Repo:        &models.Repo{Name: "test", URL: "http://oci-test/test"},
					Description: "chart description",
					Home:        "https://kubeapps.com",
					Keywords:    []string{"helm"},
					Maintainers: []chart.Maintainer{{Name: "Bitnami", Email: "containers@bitnami.com"}},
					Sources:     []string{"https://github.com/vmware-tanzu/kubeapps"},
					Icon:        "https://logo.png",
					Category:    "Infrastructure",
					ChartVersions: []models.ChartVersion{
						{
							Version:    "1.0.0",
							AppVersion: "2.0.0",
							Digest:     "123",
							URLs:       []string{"https://github.com/vmware-tanzu/kubeapps"},
						},
					},
				},
			},
			false,
		},
		{
			"Retrieve other files",
			"kubeapps",
			[]tartest.TarballFile{
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/kubeapps",
					Name:        "kubeapps",
					Repo:        &models.Repo{Name: "test", URL: "http://oci-test/test"},
					Maintainers: []chart.Maintainer{},
					ChartVersions: []models.ChartVersion{
						{
							Digest: "123",
							Readme: "chart readme",
							Values: "chart values",
							Schema: "chart schema",
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
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "repo%2Fkubeapps",
					Repo:        &models.Repo{Name: "test", URL: "http://oci-test/"},
					Maintainers: []chart.Maintainer{},
					ChartVersions: []models.ChartVersion{
						{
							Digest: "123",
							Readme: "chart readme",
							Values: "chart values",
							Schema: "chart schema",
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
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.0.0", "1.1.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "repo%2Fkubeapps",
					Repo:        &models.Repo{Name: "test", URL: "http://oci-test/"},
					Maintainers: []chart.Maintainer{},
					ChartVersions: []models.ChartVersion{
						{
							Digest: "123",
							Readme: "chart readme",
							Values: "chart values",
							Schema: "chart schema",
						},
						{
							Digest: "123",
							Readme: "chart readme",
							Values: "chart values",
							Schema: "chart schema",
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
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			[]string{"1.1.0", "1.0.0"},
			[]models.Chart{
				{
					ID:          "test/repo%2Fkubeapps",
					Name:        "repo%2Fkubeapps",
					Repo:        &models.Repo{Name: "test", URL: "http://oci-test/"},
					Maintainers: []chart.Maintainer{},
					ChartVersions: []models.ChartVersion{
						{
							Digest: "123",
							Readme: "chart readme",
							Values: "chart values",
							Schema: "chart schema",
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
			url, _ := parseRepoURL("http://oci-test")

			tags := map[string]string{}
			for _, tag := range tt.tags {
				tags[fmt.Sprintf("/v2/%s/manifests/%s", tt.chartName, tag)] = `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:123","size":665}}`
			}
			chartsRepo := OCIRegistry{
				repositories: []string{tt.chartName},
				RepoInternal: &models.RepoInternal{Name: tt.expected[0].Repo.Name, URL: tt.expected[0].Repo.URL},
				tags: map[string]TagList{
					tt.chartName: {Name: fmt.Sprintf("test/%s", tt.chartName), Tags: tt.tags},
				},
				puller: &helmfake.OCIPuller{
					Content:  content,
					Checksum: "123",
				},
				ociCli: &ociAPICli{
					url: url,
					netClient: &goodOCIAPIHTTPClient{
						responseByPath: tags,
					},
				},
			}
			charts, err := chartsRepo.Charts(tt.shallow)
			assert.NoError(t, err)
			if !cmp.Equal(charts, tt.expected) {
				t.Errorf("Unexpected result %v", cmp.Diff(charts, tt.expected))
			}
		})
	}

	t.Run("FetchFiles - It returns the stored files", func(t *testing.T) {
		files := map[string]string{
			models.ValuesKey: "values text",
			models.ReadmeKey: "readme text",
			models.SchemaKey: "schema text",
		}
		repo := OCIRegistry{}
		result, err := repo.FetchFiles("", models.ChartVersion{
			Values: files["values"],
			Readme: files["readme"],
			Schema: files["schema"],
		}, "my-user-agent", false)
		assert.NoError(t, err)
		assert.Equal(t, result, files, "expected files")
	})
}

func Test_extractFilesFromBuffer(t *testing.T) {
	tests := []struct {
		description string
		files       []tartest.TarballFile
		expected    *artifactFiles
	}{
		{
			"It should extract the important files",
			[]tartest.TarballFile{
				{Name: "Chart.yaml", Body: "chart yaml"},
				{Name: "README.md", Body: "chart readme"},
				{Name: "values.yaml", Body: "chart values"},
				{Name: "values.schema.json", Body: "chart schema"},
			},
			&artifactFiles{
				Metadata: "chart yaml",
				Readme:   "chart readme",
				Values:   "chart values",
				Schema:   "chart schema",
			},
		},
		{
			"It should ignore letter case",
			[]tartest.TarballFile{
				{Name: "Readme.md", Body: "chart readme"},
			},
			&artifactFiles{
				Readme: "chart readme",
			},
		},
		{
			"It should ignore other files",
			[]tartest.TarballFile{
				{Name: "README.md", Body: "chart readme"},
				{Name: "other.yaml", Body: "other content"},
			},
			&artifactFiles{
				Readme: "chart readme",
			},
		},
		{
			"It should handle large files",
			[]tartest.TarballFile{
				// 1MB file
				{Name: "README.md", Body: string(make([]byte, 1048577))},
			},
			&artifactFiles{
				Readme: string(make([]byte, 1048577)),
			},
		},
		{
			"It should ignore nested files",
			[]tartest.TarballFile{
				{Name: "other/README.md", Body: "bad"},
				{Name: "README.md", Body: "good"},
			},
			&artifactFiles{
				Readme: "good",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			gzw := gzip.NewWriter(w)
			tartest.CreateTestTarball(gzw, tt.files)
			gzw.Flush()

			r, err := extractFilesFromBuffer(w.Body)
			assert.NoError(t, err)
			if !cmp.Equal(r, tt.expected) {
				t.Errorf("Unexpected result %v", cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_filterCharts(t *testing.T) {
	tests := []struct {
		description string
		input       []models.Chart
		rule        apprepov1alpha1.FilterRuleSpec
		expected    []models.Chart
		expectedErr error
	}{
		{
			"should filter a chart",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1", Variables: map[string]string{"$var1": "foo"},
			},
			[]models.Chart{
				{Name: "foo"},
			},
			nil,
		},
		{
			"an invalid rule cause to return an empty set",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: "not a rule",
			},
			nil,
			fmt.Errorf(`Unable to parse jq query: unexpected token "a"`),
		},
		{
			"an invalid number of vars cause to return an empty set",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1",
			},
			nil,
			fmt.Errorf(`Unable to compile jq: variable not defined: $var1`),
		},
		{
			"the query doesn't return a boolean",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name`,
			},
			nil,
			fmt.Errorf(`Unable to convert jq result to boolean. Got: foo`),
		},
		{
			"matches without vars",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name == "foo"`,
			},
			[]models.Chart{
				{Name: "foo"},
			},
			nil,
		},
		{
			"filters a maintainer name",
			[]models.Chart{
				{Name: "foo", Maintainers: []chart.Maintainer{{Name: "Bitnami"}}},
				{Name: "bar", Maintainers: []chart.Maintainer{{Name: "Hackers"}}},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".maintainers | any(.name == $var1)", Variables: map[string]string{"$var1": "Bitnami"},
			},
			[]models.Chart{
				{Name: "foo", Maintainers: []chart.Maintainer{{Name: "Bitnami"}}},
			},
			nil,
		},
		{
			"excludes a value",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: ".name == $var1 | not", Variables: map[string]string{"$var1": "foo"},
			},
			[]models.Chart{
				{Name: "bar"},
			},
			nil,
		},
		{
			"matches against a regex",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{
				JQ: `.name | test($var1)`, Variables: map[string]string{"$var1": ".*oo.*"},
			},
			[]models.Chart{
				{Name: "foo"},
			},
			nil,
		},
		{
			"ignores an empty rule",
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			apprepov1alpha1.FilterRuleSpec{},
			[]models.Chart{
				{Name: "foo"},
				{Name: "bar"},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			res, err := filterCharts(tt.input, &tt.rule)
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
		input       []models.Chart
		expected    []models.Chart
	}{
		{
			"chart with encoded spaces in id",
			[]models.Chart{
				{ID: "foo%20bar"},
			},
			[]models.Chart{
				{ID: "foo bar"},
			},
		},
		{
			"chart with encoded spaces in name",
			[]models.Chart{
				{Name: "foo%20bar"},
			},
			[]models.Chart{
				{Name: "foo bar"},
			},
		},
		{
			"chart with mixed encoding in name",
			[]models.Chart{
				{Name: "test/foo%20bar"},
			},
			[]models.Chart{
				{Name: "test/foo bar"},
			},
		},
		{
			"chart with no encoding nor spaces",
			[]models.Chart{
				{Name: "test/foobar"},
			},
			[]models.Chart{
				{Name: "test/foobar"},
			},
		},
		{
			"chart with unencoded spaces",
			[]models.Chart{
				{Name: "test/foo bar"},
			},
			[]models.Chart{
				{Name: "test/foo bar"},
			},
		},
		{
			"chart with encoded chars in name",
			[]models.Chart{
				{Name: "foo%23bar%2ebar"},
			},
			[]models.Chart{
				{Name: "foo#bar.bar"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			res := unescapeChartsData(tt.input)
			if !cmp.Equal(res, tt.expected) {
				t.Errorf("Unexpected result: %v", cmp.Diff(res, tt.expected))
			}
		})
	}
}

func TestHelmRepoAppliesUnescape(t *testing.T) {
	repo := &models.RepoInternal{Name: "test", Namespace: "repo-namespace", URL: "http://testrepo.com"}
	expectedRepo := &models.Repo{Name: repo.Name, Namespace: repo.Namespace, URL: repo.URL}
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
	helmRepo := &HelmRepo{
		content:      []byte(repoIndexYAML),
		RepoInternal: repo,
	}
	t.Run("Helm repo applies unescaping to chart data", func(t *testing.T) {
		charts, _ := helmRepo.Charts(false)
		if !cmp.Equal(charts, expectedCharts) {
			t.Errorf("Unexpected result: %v", cmp.Diff(charts, expectedCharts))
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
