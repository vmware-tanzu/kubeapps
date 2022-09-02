// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
)

var validRepoIndexYAMLBytes, _ = os.ReadFile("testdata/valid-index.yaml")
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
			assert.Error(t, err, tt.name)
		})
	}

	t.Run("valid", func(t *testing.T) {
		index, err := parseRepoIndex([]byte(validRepoIndexYAML))
		assert.NoError(t, err)
		assert.Equal(t, len(index.Entries), 2, "number of charts")
		assert.Equal(t, index.Entries["acs-engine-autoscaler"][0].Name, "acs-engine-autoscaler", "chart version populated")
	})
}

func Test_chartsFromIndex(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	charts, err := ChartsFromIndex([]byte(validRepoIndexYAML), r, false)
	assert.NoError(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")

	indexWithDeprecated := validRepoIndexYAML + `
  deprecated-chart:
  - name: deprecated-chart
    deprecated: true`
	charts, err = ChartsFromIndex([]byte(indexWithDeprecated), r, false)
	assert.NoError(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")
	assert.Equal(t, len(charts[1].ChartVersions), 2, "number of versions")
}

func Test_shallowChartsFromIndex(t *testing.T) {
	r := &models.Repo{Name: "test", URL: "http://testrepo.com"}
	charts, err := ChartsFromIndex([]byte(validRepoIndexYAML), r, true)
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, len(charts), 2, "number of charts")
	assert.Equal(t, len(charts[1].ChartVersions), 1, "number of versions")
}
