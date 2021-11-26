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
package tarutil

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/kubeapps/kubeapps/pkg/tarutil/test"
	"github.com/stretchr/testify/assert"
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
		rand.Read(b)
		r := bytes.NewReader(b)
		tarf := tar.NewReader(r)
		values := "values"
		files, err := ExtractFilesFromTarball(map[string]string{values: "file2.txt"}, tarf)
		assert.Error(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, len(files), 0, "file body")
	})
}
