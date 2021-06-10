/*
Copyright © 2021 VMware
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
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/ghodss/yaml"
	"github.com/kubeapps/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	helmrepo "k8s.io/helm/pkg/repo"
	log "k8s.io/klog/v2"
)

const (
	readme = "readme"
)

//
// TODO some of this functionality already exists in asset-syncer but is private
// so it needs to be re-packaged so that it can be re-used
//
func getHelmIndexFileFromURL(indexURL string) (*helmrepo.IndexFile, error) {
	log.Infof("+getHelmIndexFileFromURL(%s) 1", indexURL)
	// Get the response bytes from the url
	response, err := http.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.FailedPrecondition, "received non OK response code: [%d]", response.StatusCode)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var index helmrepo.IndexFile
	err = yaml.Unmarshal(contents, &index)
	if err != nil {
		return nil, err
	}
	index.SortEntries()
	log.Infof("-getHelmIndexFileFromURL(%s)", indexURL)
	return &index, nil
}

func fetchMetaFromChartTarball(name string, chartTarballURL string) (map[string]string, error) {
	// Get the response bytes from the url
	response, err := http.Get(chartTarballURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.FailedPrecondition, "received non OK response code: [%d]", response.StatusCode)
	}

	// We read the whole chart into memory, this should be okay since the chart
	// tarball needs to be small enough to fit into a GRPC call (Tiller
	// requirement)
	gzf, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer gzf.Close()

	tarf := tar.NewReader(gzf)

	// decode escaped characters
	// ie., "foo%2Fbar" should return "foo/bar"
	decodedName, err := url.PathUnescape(name)
	if err != nil {
		return nil, err
	}

	// get last part of the name
	// ie., "foo/bar" should return "bar"
	fixedName := path.Base(decodedName)

	filenames := map[string]string{
		readme: fixedName + "/README.md",
	}

	files, err := tarutil.ExtractFilesFromTarball(filenames, tarf)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		readme: files[readme],
	}, nil
}
