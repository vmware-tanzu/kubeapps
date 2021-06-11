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
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	helmrepo "k8s.io/helm/pkg/repo"
	log "k8s.io/klog/v2"
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
