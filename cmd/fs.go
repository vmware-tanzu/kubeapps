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

package cmd

import (
	"bytes"
	_ "github.com/kubeapps/installer/generated/statik"
	"github.com/rakyll/statik/fs"
	"io"
	"log"
	"net/http"
)

var statikFS http.FileSystem

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatalf("ERROR initializing statikFS")
	}
}

func fsGetFile(fname string) (string, error) {
	buf := bytes.NewBuffer(nil)
	statik_file, err := statikFS.Open(fname)
	if err != nil {
		log.Fatalf("ERROR: Static file '%s' not found", fname)
		return "", err
	}
	io.Copy(buf, statik_file)
	content := string(buf.Bytes())
	return content, nil
}
