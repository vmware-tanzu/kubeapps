/*
Copyright (c) 2017 Bitnami

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
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"
	"github.com/kubeapps/common/datastore"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	helmrepo "k8s.io/helm/pkg/repo"
)

const (
	chartCollection      = "charts"
	chartFilesCollection = "files"
)

type importChartFilesJob struct {
	Name         string
	Repo         repo
	ChartVersion chartVersion
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var dbSession datastore.Session

var netClient httpClient = &http.Client{
	Timeout: time.Second * 10,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [REPO NAME] [REPO URL]\n", os.Args[0])
		flag.PrintDefaults()
	}
	dbURL := flag.String("mongo-url", "localhost", "MongoDB URL (see https://godoc.org/labix.org/v2/mgo#Dial for format)")
	dbName := flag.String("mongo-database", "charts", "MongoDB database")
	dbUsername := flag.String("mongo-user", "", "MongoDB user")
	dbPassword := os.Getenv("MONGO_PASSWORD")
	debug := flag.Bool("debug", false, "verbose logging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	mongoConfig := datastore.Config{URL: *dbURL, Database: *dbName, Username: *dbUsername, Password: dbPassword}
	var err error
	dbSession, err = datastore.NewSession(mongoConfig)
	if err != nil {
		log.WithFields(log.Fields{"host": *dbURL}).Fatal(err)
	}

	if err := syncRepo(flag.Arg(0), flag.Arg(1)); err != nil {
		log.WithError(err).Error("sync failed")
		os.Exit(1)
	}
}

// Syncing is performed in the following steps:
// 1. Update database to match chart metadata from index
// 2. Concurrently process icons for charts (concurrently)
// 3. Concurrently process the README and values.yaml for the latest chart version of each chart
// 4. Concurrently process READMEs and values.yaml for historic chart versions
//
// These steps are processed in this way to ensure relevant chart data is
// imported into the database as fast as possible. E.g. we want all icons for
// charts before fetching readmes for each chart and version pair.
func syncRepo(repoName, repoURL string) error {
	url, err := url.ParseRequestURI(repoURL)
	if err != nil {
		log.WithFields(log.Fields{"url": repoURL}).WithError(err).Error("failed to parse URL")
		return err
	}

	index, err := fetchRepoIndex(url)
	if err != nil {
		return err
	}

	charts := chartsFromIndex(index, repo{Name: repoName, URL: repoURL})
	err = importCharts(charts)
	if err != nil {
		return err
	}

	// Process 10 charts at a time
	numWorkers := 10
	iconJobs := make(chan chart, numWorkers)
	chartFilesJobs := make(chan importChartFilesJob, numWorkers)
	var wg sync.WaitGroup

	log.Debugf("starting %d workers", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go importWorker(&wg, iconJobs, chartFilesJobs)
	}

	// Enqueue jobs to process chart icons
	for _, c := range charts {
		iconJobs <- c
	}
	// Close the iconJobs channel to signal the worker pools to move on to the
	// chart files jobs
	close(iconJobs)

	// Iterate through the list of charts and enqueue the latest chart version to
	// be processed. Append the rest of the chart versions to a list to be
	// enqueued later
	var toEnqueue []importChartFilesJob
	for _, c := range charts {
		chartFilesJobs <- importChartFilesJob{c.Name, c.Repo, c.ChartVersions[0]}
		for _, cv := range c.ChartVersions[1:] {
			toEnqueue = append(toEnqueue, importChartFilesJob{c.Name, c.Repo, cv})
		}
	}

	// Enqueue all the remaining chart versions
	for _, cfj := range toEnqueue {
		chartFilesJobs <- cfj
	}
	// Close the chartFilesJobs channel to signal the worker pools that there are
	// no more jobs to process
	close(chartFilesJobs)

	// Wait for the worker pools to finish processing
	wg.Wait()

	return nil
}

func fetchRepoIndex(repoURL *url.URL) (*helmrepo.IndexFile, error) {
	// use a copy of the URL struct so we don't modify the original
	indexURL := *repoURL
	indexURL.Path = path.Join(indexURL.Path, "index.yaml")
	req, err := http.NewRequest("GET", indexURL.String(), nil)
	req.Header.Set("User-Agent", userAgent)
	if err != nil {
		log.WithFields(log.Fields{"url": req.URL.String()}).WithError(err).Error("could not build repo index request")
		return nil, err
	}
	res, err := netClient.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"url": req.URL.String()}).WithError(err).Error("error requesting repo index")
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"url": req.URL.String(), "status": res.StatusCode}).Error("error requesting repo index, are you sure this is a chart repository?")
		return nil, errors.New("repo index request failed")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return parseRepoIndex(body)
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

func chartsFromIndex(index *helmrepo.IndexFile, r repo) []chart {
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
func newChart(entry helmrepo.ChartVersions, r repo) chart {
	var c chart
	copier.Copy(&c, entry[0])
	copier.Copy(&c.ChartVersions, entry)
	c.Repo = r
	c.ID = fmt.Sprintf("%s/%s", r.Name, c.Name)
	return c
}

func importCharts(charts []chart) error {
	var pairs []interface{}
	var chartIDs []string
	for _, c := range charts {
		chartIDs = append(chartIDs, c.ID)
		// charts to upsert - pair of selector, chart
		pairs = append(pairs, bson.M{"_id": c.ID}, c)
	}

	db, closer := dbSession.DB()
	defer closer()
	bulk := db.C(chartCollection).Bulk()

	// Upsert pairs of selectors, charts
	bulk.Upsert(pairs...)

	// Remove charts no longer existing in index
	bulk.RemoveAll(bson.M{
		"_id": bson.M{
			"$nin": chartIDs,
		},
		"repo.name": charts[0].Repo.Name,
	})

	_, err := bulk.Run()
	return err
}

func importWorker(wg *sync.WaitGroup, icons <-chan chart, chartFiles <-chan importChartFilesJob) {
	defer wg.Done()
	for c := range icons {
		log.WithFields(log.Fields{"name": c.Name}).Debug("importing icon")
		if err := fetchAndImportIcon(c); err != nil {
			log.WithFields(log.Fields{"name": c.Name}).WithError(err).Error("failed to import icon")
		}
	}
	for j := range chartFiles {
		log.WithFields(log.Fields{"name": j.Name, "version": j.ChartVersion.Version}).Debug("importing readme and values")
		if err := fetchAndImportFiles(j.Name, j.Repo, j.ChartVersion); err != nil {
			log.WithFields(log.Fields{"name": j.Name, "version": j.ChartVersion.Version}).WithError(err).Error("failed to import files")
		}
	}
}

func fetchAndImportIcon(c chart) error {
	if c.Icon == "" {
		log.WithFields(log.Fields{"name": c.Name}).Info("icon not found")
		return nil
	}

	req, err := http.NewRequest("GET", c.Icon, nil)
	req.Header.Set("User-Agent", userAgent)
	if err != nil {
		return err
	}

	res, err := netClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", res.StatusCode, c.Icon)
	}

	orig, err := imaging.Decode(res.Body)
	if err != nil {
		return err
	}

	// TODO: make this configurable?
	icon := imaging.Fit(orig, 160, 160, imaging.Lanczos)

	var b bytes.Buffer
	imaging.Encode(&b, icon, imaging.PNG)

	db, closer := dbSession.DB()
	defer closer()
	return db.C(chartCollection).UpdateId(c.ID, bson.M{"$set": bson.M{"raw_icon": b.Bytes()}})
}

func fetchAndImportFiles(name string, r repo, cv chartVersion) error {
	chartFilesID := fmt.Sprintf("%s/%s-%s", r.Name, name, cv.Version)
	db, closer := dbSession.DB()
	defer closer()
	if err := db.C(chartFilesCollection).FindId(chartFilesID).One(&chartFiles{}); err == nil {
		log.WithFields(log.Fields{"name": name, "version": cv.Version}).Debug("skipping existing files")
		return nil
	}
	log.WithFields(log.Fields{"name": name, "version": cv.Version}).Debug("fetching files")

	url := chartTarballURL(r, cv)
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", userAgent)
	if err != nil {
		return err
	}

	res, err := netClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// We read the whole chart into memory, this should be okay since the chart
	// tarball needs to be small enough to fit into a GRPC call (Tiller
	// requirement)
	gzf, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer gzf.Close()

	tarf := tar.NewReader(gzf)

	readmeFileName := name + "/README.md"
	valuesFileName := name + "/values.yaml"
	fileNames := []string{valuesFileName, readmeFileName}

	files, err := getFiles(cv, name, fileNames, tarf)
	values := files[0]
	readme := files[1]

	db.C(chartFilesCollection).Insert(chartFiles{chartFilesID, readme, values})

	return nil
}

func getFiles(cv chartVersion, name string, filenames []string, tarf *tar.Reader) ([]string, error) {
	var files []string

	for _, filename := range filenames {
		file, err := extractFileFromTarball(filename, tarf)
		if err != nil && !strings.Contains(err.Error(), "file not found") {
			return nil, err
		}

		if file == "" {
			log.WithFields(log.Fields{"name": name, "version": cv.Version}).Info(filename + " not found")
		}

		files = append(files, file)
	}

	return files, nil
}

func chartTarballURL(r repo, cv chartVersion) string {
	source := cv.URLs[0]
	if _, err := url.ParseRequestURI(source); err != nil {
		// If the chart URL is not absolute, join with repo URL. It's fine if the
		// URL we build here is invalid as we can catch this error when actually
		// making the request
		u, _ := url.Parse(r.URL)
		u.Path = path.Join(u.Path, source)
		return u.String()
	}
	return source
}

func extractFileFromTarball(filename string, tarf *tar.Reader) (string, error) {
	for {
		header, err := tarf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Name == filename {
			var b bytes.Buffer
			io.Copy(&b, tarf)
			return string(b.Bytes()), nil
		}
	}
	return "", fmt.Errorf("%s file not found", filename)
}
