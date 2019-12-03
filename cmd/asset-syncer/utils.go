/*
Copyright (c) 2018 The Helm Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"
	"github.com/kubeapps/common/datastore"
	log "github.com/sirupsen/logrus"
	helmrepo "k8s.io/helm/pkg/repo"
)

const (
	defaultTimeoutSeconds = 10
	additionalCAFile      = "/usr/local/share/ca-certificates/ca.crt"
)

type importChartFilesJob struct {
	Name         string
	Repo         *repo
	ChartVersion chartVersion
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var netClient httpClient = &http.Client{}

func parseRepoURL(repoURL string) (*url.URL, error) {
	repoURL = strings.TrimSpace(repoURL)
	return url.ParseRequestURI(repoURL)
}

func init() {
	var err error
	netClient, err = initNetClient(additionalCAFile)
	if err != nil {
		log.Fatal(err)
	}
}

type assetManager interface {
	Delete(repo string) error
	Sync(charts []chart) error
	RepoAlreadyProcessed(repoName, checksum string) bool
	UpdateLastCheck(repoName, checksum string, now time.Time) error
	Init() error
	Close() error
}

func newManager(databaseType, databaseURL, databaseName, databaseUser, databasePassword string) (assetManager, error) {
	if databaseType == "mongodb" {
		mongoConfig := datastore.Config{URL: databaseURL, Database: databaseName, Username: databaseUser, Password: databasePassword}
		return &mongodbAssetManager{mongoConfig, nil}, nil
	} else if databaseType == "postgresql" {
		url := strings.Split(databaseURL, ":")
		if len(url) != 2 {
			return nil, fmt.Errorf("Can't parse database URL: %s", databaseURL)
		}
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			url[0], url[1], databaseUser, databasePassword, databaseName,
		)
		return &postgresAssetManager{connStr, nil}, nil
	} else {
		return nil, fmt.Errorf("Unsupported database type %s", databaseType)
	}
}

func getSha256(src []byte) (string, error) {
	f := bytes.NewReader(src)
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getRepo(name, repoURL, authorizationHeader string) (*repo, error) {
	url, err := parseRepoURL(repoURL)
	if err != nil {
		log.WithFields(log.Fields{"url": repoURL}).WithError(err).Error("failed to parse URL")
		return nil, err
	}

	repoBytes, err := fetchRepoIndex(url.String(), authorizationHeader)
	if err != nil {
		return nil, err
	}

	repoChecksum, err := getSha256(repoBytes)
	if err != nil {
		return nil, err
	}

	return &repo{Name: name, URL: url.String(), AuthorizationHeader: authorizationHeader, Checksum: repoChecksum, Content: repoBytes}, nil
}

func fetchRepoIndex(url, authHeader string) ([]byte, error) {
	indexURL, err := parseRepoURL(url)
	if err != nil {
		log.WithFields(log.Fields{"url": url}).WithError(err).Error("failed to parse URL")
		return nil, err
	}
	indexURL.Path = path.Join(indexURL.Path, "index.yaml")
	req, err := http.NewRequest("GET", indexURL.String(), nil)
	if err != nil {
		log.WithFields(log.Fields{"url": req.URL.String()}).WithError(err).Error("could not build repo index request")
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent())
	if len(authHeader) > 0 {
		req.Header.Set("Authorization", authHeader)
	}

	res, err := netClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		log.WithFields(log.Fields{"url": req.URL.String()}).WithError(err).Error("error requesting repo index")
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"url": req.URL.String(), "status": res.StatusCode}).Error("error requesting repo index, are you sure this is a chart repository?")
		return nil, errors.New("repo index request failed")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func parseRepoIndex(body []byte) (*helmrepo.IndexFile, error) {
	var index helmrepo.IndexFile
	err := yaml.Unmarshal(body, &index)
	if err != nil {
		return nil, err
	}
	index.SortEntries()
	return &index, nil
}

func chartsFromIndex(index *helmrepo.IndexFile, r *repo) []chart {
	var charts []chart
	for _, entry := range index.Entries {
		if entry[0].GetDeprecated() {
			log.WithFields(log.Fields{"name": entry[0].GetName()}).Info("skipping deprecated chart")
			continue
		}
		charts = append(charts, newChart(entry, r))
	}
	return charts
}

// Takes an entry from the index and constructs a database representation of the
// object.
func newChart(entry helmrepo.ChartVersions, r *repo) chart {
	var c chart
	copier.Copy(&c, entry[0])
	copier.Copy(&c.ChartVersions, entry)
	c.Repo = r
	c.ID = fmt.Sprintf("%s/%s", r.Name, c.Name)
	return c
}

func extractFilesFromTarball(filenames []string, tarf *tar.Reader) (map[string]string, error) {
	ret := make(map[string]string)
	for {
		header, err := tarf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ret, err
		}

		for _, f := range filenames {
			if strings.EqualFold(header.Name, f) {
				var b bytes.Buffer
				io.Copy(&b, tarf)
				ret[f] = string(b.Bytes())
				break
			}
		}
	}
	return ret, nil
}

func chartTarballURL(r *repo, cv chartVersion) string {
	source := cv.URLs[0]
	if _, err := parseRepoURL(source); err != nil {
		// If the chart URL is not absolute, join with repo URL. It's fine if the
		// URL we build here is invalid as we can catch this error when actually
		// making the request
		u, _ := url.Parse(r.URL)
		u.Path = path.Join(u.Path, source)
		return u.String()
	}
	return source
}

func initNetClient(additionalCA string) (*http.Client, error) {
	// Get the SystemCertPool, continue with an empty pool on error
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}

	// If additionalCA exists, load it
	if _, err := os.Stat(additionalCA); !os.IsNotExist(err) {
		certs, err := ioutil.ReadFile(additionalCA)
		if err != nil {
			return nil, fmt.Errorf("Failed to append %s to RootCAs: %v", additionalCA, err)
		}

		// Append our cert to the system pool
		if ok := caCertPool.AppendCertsFromPEM(certs); !ok {
			return nil, fmt.Errorf("Failed to append %s to RootCAs", additionalCA)
		}
	}

	// Return Transport for testing purposes
	return &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}, nil
}
