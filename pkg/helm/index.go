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
package helm

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	helmrepo "k8s.io/helm/pkg/repo"
	log "k8s.io/klog/v2"
)

func parseRepoIndex(contents []byte) (*helmrepo.IndexFile, error) {
	var index helmrepo.IndexFile
	err := yaml.Unmarshal(contents, &index)
	if err != nil {
		return nil, err
	}
	index.SortEntries()
	return &index, nil
}

// Takes an entry from the index and constructs a model representation of the
// object.
func newChart(entry helmrepo.ChartVersions, r *models.Repo, shallow bool) models.Chart {
	var c models.Chart
	copier.Copy(&c, entry[0])
	if shallow {
		copier.Copy(&c.ChartVersions, []helmrepo.ChartVersion{*entry[0]})
	} else {
		copier.Copy(&c.ChartVersions, entry)
	}
	c.Repo = r
	c.Name = url.PathEscape(c.Name) // escaped chart name eg. foo/bar becomes foo%2Fbar
	c.ID = fmt.Sprintf("%s/%s", r.Name, c.Name)
	c.Category = entry[0].Annotations["category"]
	return c
}

//
// ChartsFromIndex receives an array of bytes containing the contents of index.yaml from a helm repo and returns
// all Chart models from that index. The shallow flag controls whether only the latest version of the charts is returned
// or all versions
//
func ChartsFromIndex(contents []byte, r *models.Repo, shallow bool) ([]models.Chart, error) {
	var charts []models.Chart
	index, err := parseRepoIndex(contents)
	if err != nil {
		return []models.Chart{}, err
	}
	for key, entry := range index.Entries {
		// note that 'entry' itself is an array of chart versions
		// after index.SortEntires() call, it looks like there is only one entry per package,
		// and entry[0] should be the most recent chart version, e.g. Name: "mariadb" Version: "9.3.12"
		// while the rest of the elements in the entry array keep track of previous chart versions, e.g.
		// "mariadb" version "9.3.11", "9.3.10", etc. For entry "mariadb", bitnami catalog has
		// almost 200 chart versions going all the way back many years to version "2.1.4".
		// So for now, let's just keep track of the latest, not to overwhelm the caller with
		// all these outdated versions

		// skip if the entry is empty
		if len(entry) < 1 {
			log.Infof("skipping chart: [%s]", key)
			continue
		}

		if entry[0].GetDeprecated() {
			log.Infof("skipping deprecated chart: [%s]", entry[0].Name)
			continue
		}
		charts = append(charts, newChart(entry, r, shallow))
	}
	sort.Slice(charts, func(i, j int) bool { return charts[i].ID < charts[j].ID })
	return charts, nil
}
