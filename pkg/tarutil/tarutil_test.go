// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package tarutil

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"io"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	chart "github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil/test"
)

func Test_extractFilesFromTarball(t *testing.T) {
	tests := []struct {
		name     string
		files    []test.TarballFile
		filename string
		want     string
	}{
		{
			name:     "file",
			files:    []test.TarballFile{{Name: "file.txt", Body: "best file ever"}},
			filename: "file.txt",
			want:     "best file ever",
		},
		{
			name:     "multiple file tarball",
			files:    []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "file2.txt", Body: "worst file ever"}},
			filename: "file2.txt",
			want:     "worst file ever",
		},
		{
			name:     "file in dir with parent dir specified",
			files:    []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "test/file2.txt", Body: "worst file ever"}},
			filename: "test/file2.txt",
			want:     "worst file ever",
		},
		{
			name:     "file in dir without parent dir specified",
			files:    []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "test/file2.txt", Body: "worst file ever"}},
			filename: "file2.txt",
			want:     "worst file ever",
		},
		{
			name:     "filename ignore case",
			files:    []test.TarballFile{{Name: "Readme.md", Body: "# readme for chart"}, {Name: "values.yaml", Body: "key: value"}},
			filename: "README.md",
			want:     "# readme for chart",
		},
		{
			name:     "different paths to same base filename",
			files:    []test.TarballFile{{Name: "test/readme.md", Body: "# readme for chart"}},
			filename: "other/readme.md",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			test.CreateTestTarball(&b, tt.files)
			r := bytes.NewReader(b.Bytes())
			tarf := tar.NewReader(r)
			files, err := ExtractFilesFromTarball(map[string]string{tt.filename: tt.filename}, map[string]*regexp.Regexp{}, tarf)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, files[tt.filename], "file body")
		})
	}

	t.Run("extract multiple files", func(t *testing.T) {
		var b bytes.Buffer
		tFiles := []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "file2.txt", Body: "worst file ever"}}
		test.CreateTestTarball(&b, tFiles)
		r := bytes.NewReader(b.Bytes())
		tarf := tar.NewReader(r)
		files, err := ExtractFilesFromTarball(map[string]string{tFiles[0].Name: tFiles[0].Name, tFiles[1].Name: tFiles[1].Name}, map[string]*regexp.Regexp{}, tarf)
		assert.NoError(t, err)
		assert.Equal(t, len(files), 2, "matches")
		for _, f := range tFiles {
			assert.Equal(t, files[f.Name], f.Body, "file body")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		var b bytes.Buffer
		test.CreateTestTarball(&b, []test.TarballFile{{Name: "file.txt", Body: "best file ever"}})
		r := bytes.NewReader(b.Bytes())
		tarf := tar.NewReader(r)
		name := "file2.txt"
		files, err := ExtractFilesFromTarball(map[string]string{name: name}, map[string]*regexp.Regexp{}, tarf)
		assert.NoError(t, err)
		assert.Equal(t, files[name], "", "file body")
	})

	t.Run("not a tarball", func(t *testing.T) {
		b := make([]byte, 4)
		_, err := rand.Read(b)
		assert.NoError(t, err)
		r := bytes.NewReader(b)
		tarf := tar.NewReader(r)
		values := "values"
		files, err := ExtractFilesFromTarball(map[string]string{values: "file2.txt"}, map[string]*regexp.Regexp{}, tarf)
		assert.Error(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, len(files), 0, "file body")
	})

	t.Run("extract files matching regex", func(t *testing.T) {
		var b bytes.Buffer
		tFiles := []test.TarballFile{
			{Name: "file.txt", Body: "best file ever"},
			{Name: "file2.txt", Body: "worst file ever"},
		}
		test.CreateTestTarball(&b, tFiles)
		r := bytes.NewReader(b.Bytes())
		tarf := tar.NewReader(r)
		files, err := ExtractFilesFromTarball(
			map[string]string{},
			map[string]*regexp.Regexp{
				"file$num.txt": regexp.MustCompile(`file(?P<num>\d*)\.txt`),
			}, tarf)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(files), "incorrect number of files found")
		for _, f := range tFiles {
			assert.Equal(t, f.Body, files[f.Name], "file content did not match")
		}
	})

	t.Run("extract files matching either name or regex, not both", func(t *testing.T) {
		var b bytes.Buffer
		tFiles := []test.TarballFile{
			{Name: "file.txt", Body: "best file ever"},
			{Name: "file2.txt", Body: "worst file ever"},
		}
		test.CreateTestTarball(&b, tFiles)
		r := bytes.NewReader(b.Bytes())
		tarf := tar.NewReader(r)
		files, err := ExtractFilesFromTarball(
			map[string]string{"file.txt": "file.txt"},
			map[string]*regexp.Regexp{
				"file$num.txt": regexp.MustCompile(`file(?P<num>\d*)\.txt`),
			}, tarf)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(files), "incorrect number of files found")
		for _, f := range tFiles {
			assert.Equal(t, f.Body, files[f.Name], "file content did not match")
		}
	})
}

func Test_FetchChartDetailFromTarball(t *testing.T) {
	pkgName := "wordpress"
	tests := []struct {
		name string
		pkg  []test.TarballFile
		want map[string]string
	}{
		{
			name: "package without any expected files",
			pkg: []test.TarballFile{
				{Name: path.Join(pkgName, "somefile.txt"), Body: "best file ever"},
				{Name: path.Join(pkgName, "somefile2.txt"), Body: "worst file ever"},
			},
			want: map[string]string{},
		},
		{
			name: "package with some chart data",
			pkg: []test.TarballFile{
				{Name: path.Join(pkgName, "somefile.txt"), Body: "best file ever"},
				{Name: path.Join(pkgName, "values.yaml"), Body: "foo: bar"},
			},
			want: map[string]string{
				chart.DefaultValuesKey: "foo: bar",
			},
		},
		{
			name: "package with all chart data",
			pkg: []test.TarballFile{
				{Name: path.Join(pkgName, "values.yaml"), Body: "foo: bar"},
				{Name: path.Join(pkgName, "README.md"), Body: "readme text"},
				{Name: path.Join(pkgName, "values.schema.json"), Body: "{}"},
				{Name: path.Join(pkgName, "Chart.yaml"), Body: "Name: wordpress"},
			},
			want: map[string]string{
				chart.ChartYamlKey:     "Name: wordpress",
				chart.ReadmeKey:        "readme text",
				chart.SchemaKey:        "{}",
				chart.DefaultValuesKey: "foo: bar",
			},
		},
		{
			name: "package with additional default values",
			pkg: []test.TarballFile{
				{Name: path.Join(pkgName, "values.yaml"), Body: "foo: bar"},
				{Name: path.Join(pkgName, "README.md"), Body: "readme text"},
				{Name: path.Join(pkgName, "values.schema.json"), Body: "{}"},
				{Name: path.Join(pkgName, "Chart.yaml"), Body: "Name: wordpress"},
				{Name: path.Join(pkgName, "values-production.yaml"), Body: "foo: prod-bar"},
				{Name: path.Join(pkgName, "values-scenario-a.yaml"), Body: "scenario: a"},
			},
			want: map[string]string{
				chart.ChartYamlKey:     "Name: wordpress",
				chart.ReadmeKey:        "readme text",
				chart.SchemaKey:        "{}",
				chart.DefaultValuesKey: "foo: bar",
				"values-production":    "foo: prod-bar",
				"values-scenario-a":    "scenario: a",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			test.CreateTestTgz(&b, tt.pkg)

			r := bytes.NewReader(b.Bytes())

			files, err := FetchChartDetailFromTarball(r)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, files)
		})
	}
}
