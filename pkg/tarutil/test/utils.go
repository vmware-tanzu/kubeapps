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
package test

import (
	"archive/tar"
	"io"

	log "github.com/sirupsen/logrus"
)

type TarballFile struct {
	Name, Body string
}

//
// utility used by unit test code to create tarball files from specified array of name/contents pairs
//
func CreateTestTarball(w io.Writer, files []TarballFile) {
	// Create a new tar archive.
	tarw := tar.NewWriter(w)

	// Add files to the archive.
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tarw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tarw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}
	// Make sure to check the error on Close.
	if err := tarw.Close(); err != nil {
		log.Fatal(err)
	}
}
