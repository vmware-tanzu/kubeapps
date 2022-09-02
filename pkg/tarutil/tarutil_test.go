// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package tarutil

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil/test"
)

func Test_extractFilesFromTarball(t *testing.T) {
	tests := []struct {
		name     string
		files    []test.TarballFile
		filename string
		want     string
	}{
		{"file", []test.TarballFile{{Name: "file.txt", Body: "best file ever"}}, "file.txt", "best file ever"},
		{"multiple file tarball", []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "file2.txt", Body: "worst file ever"}}, "file2.txt", "worst file ever"},
		{"file in dir", []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "test/file2.txt", Body: "worst file ever"}}, "test/file2.txt", "worst file ever"},
		{"filename ignore case", []test.TarballFile{{Name: "Readme.md", Body: "# readme for chart"}, {Name: "values.yaml", Body: "key: value"}}, "README.md", "# readme for chart"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			test.CreateTestTarball(&b, tt.files)
			r := bytes.NewReader(b.Bytes())
			tarf := tar.NewReader(r)
			files, err := ExtractFilesFromTarball(map[string]string{tt.filename: tt.filename}, tarf)
			assert.NoError(t, err)
			assert.Equal(t, files[tt.filename], tt.want, "file body")
		})
	}

	t.Run("extract multiple files", func(t *testing.T) {
		var b bytes.Buffer
		tFiles := []test.TarballFile{{Name: "file.txt", Body: "best file ever"}, {Name: "file2.txt", Body: "worst file ever"}}
		test.CreateTestTarball(&b, tFiles)
		r := bytes.NewReader(b.Bytes())
		tarf := tar.NewReader(r)
		files, err := ExtractFilesFromTarball(map[string]string{tFiles[0].Name: tFiles[0].Name, tFiles[1].Name: tFiles[1].Name}, tarf)
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
		files, err := ExtractFilesFromTarball(map[string]string{name: name}, tarf)
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
		files, err := ExtractFilesFromTarball(map[string]string{values: "file2.txt"}, tarf)
		assert.Error(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, len(files), 0, "file body")
	})
}
