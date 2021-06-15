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
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
)

//
// Fetches helm chart details from a gzipped tarball
//
func FetchChartDetailFromTarball(name string, chartTarballURL string, userAgent string, authz string, netClient httpclient.Client) (map[string]string, error) {
	req, err := http.NewRequest("GET", chartTarballURL, nil)
	if err != nil {
		return nil, err
	}
	if len(userAgent) > 0 {
		req.Header.Set("User-Agent", userAgent)
	}

	if len(authz) > 0 {
		req.Header.Set("Authorization", authz)
	}

	res, err := netClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non OK response code: [%d]", res.StatusCode)
	}

	// We read the whole chart into memory, this should be okay since the chart
	// tarball needs to be small enough to fit into a GRPC call (Tiller
	// requirement)
	gzf, err := gzip.NewReader(res.Body)
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
	readmeFileName := fixedName + "/README.md"
	valuesFileName := fixedName + "/values.yaml"
	schemaFileName := fixedName + "/values.schema.json"
	filenames := map[string]string{
		chart.ValuesKey: valuesFileName,
		chart.ReadmeKey: readmeFileName,
		chart.SchemaKey: schemaFileName,
	}

	files, err := ExtractFilesFromTarball(filenames, tarf)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		chart.ValuesKey: files[chart.ValuesKey],
		chart.ReadmeKey: files[chart.ReadmeKey],
		chart.SchemaKey: files[chart.SchemaKey],
	}, nil
}

func ExtractFilesFromTarball(filenames map[string]string, tarf *tar.Reader) (map[string]string, error) {
	ret := make(map[string]string)
	for {
		header, err := tarf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ret, err
		}

		for id, f := range filenames {
			if strings.EqualFold(header.Name, f) {
				var b bytes.Buffer
				io.Copy(&b, tarf)
				ret[id] = b.String()
				break
			}
		}
	}
	return ret, nil
}
