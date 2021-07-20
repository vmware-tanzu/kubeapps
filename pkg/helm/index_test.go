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
	"io/ioutil"
	"testing"

	"github.com/arschles/assert"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
)

var validRepoIndexYAMLBytes, _ = ioutil.ReadFile("testdata/valid-index.yaml")
var validRepoIndexYAML = string(validRepoIndexYAMLBytes)

func Test_parseRepoIndex(t *testing.T) {
	tests := []struct {
		name     string
		repoYAML string
	}{
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseRepoIndex([]byte(tt.repoYAML))
			assert.ExistsErr(t, err, tt.name)
		})
	}

	t.Run("valid", func(t *testing.T) {
		index, err := parseRepoIndex([]byte(validRepoIndexYAML))
		assert.NoErr(t, err)
		assert.Equal(t, len(index.Entries), 2, "number of charts")
		assert.Equal(t, index.Entries["acs-engine-autoscaler"][0].GetName(), "acs-engine-autoscaler", "chart version populated")
	})
}

func Test_chartsFromIndex(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	charts, err := ChartsFromIndex([]byte(validRepoIndexYAML), r, false)
	assert.NoErr(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")

	indexWithDeprecated := validRepoIndexYAML + `
  deprecated-chart:
  - name: deprecated-chart
    deprecated: true`
	charts, err = ChartsFromIndex([]byte(indexWithDeprecated), r, false)
	assert.NoErr(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")
	assert.Equal(t, len(charts[1].ChartVersions), 2, "number of versions")
}

func Test_shallowChartsFromIndex(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	charts, err := ChartsFromIndex([]byte(validRepoIndexYAML), r, true)
	assert.NoErr(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")
	assert.Equal(t, len(charts[1].ChartVersions), 1, "number of versions")
}

func Test_newChart(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	index, _ := parseRepoIndex([]byte(validRepoIndexYAML))
	c := newChart(index.Entries["wordpress"], r, false)
	assert.Equal(t, c.Name, "wordpress", "correctly built")
	assert.Equal(t, len(c.ChartVersions), 2, "correctly built")
	assert.Equal(t, c.Description, "new description!", "takes chart fields from latest entry")
	assert.Equal(t, c.Repo, r, "repo set")
	assert.Equal(t, c.ID, "test/wordpress", "id set")
}

func Test_loadRepoWithEmptyCharts(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	indexWithEmptyChart := validRepoIndexYAML + `emptyChart: []`
	charts, err := ChartsFromIndex([]byte(indexWithEmptyChart), r, true)
	assert.NoErr(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")
	assert.Equal(t, len(charts[1].ChartVersions), 1, "number of versions")
}
