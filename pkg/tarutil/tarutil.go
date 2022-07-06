// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package tarutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/url"
	"path"
	"strings"

	chart "github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
)

//
// Fetches helm chart details from a gzipped tarball
//
// name is expected in format "foo/bar" or "foo%2Fbar" if url-escaped
//
func FetchChartDetailFromTarballUrl(name string, chartTarballURL string, userAgent string, authz string, netClient httpclient.Client) (map[string]string, error) {
	reqHeaders := make(map[string]string)
	if len(userAgent) > 0 {
		reqHeaders["User-Agent"] = userAgent
	}
	if len(authz) > 0 {
		reqHeaders["Authorization"] = authz
	}

	// use our "standard" http-client library
	reader, _, err := httpclient.GetStream(chartTarballURL, netClient, reqHeaders)
	if reader != nil {
		defer reader.Close()
	}

	if err != nil {
		return nil, err
	}

	return FetchChartDetailFromTarball(reader, name)
}

//
// Fetches helm chart details from a gzipped tarball
//
// name is expected in format "foo/bar" or "foo%2Fbar" if url-escaped
//
func FetchChartDetailFromTarball(reader io.Reader, name string) (map[string]string, error) {
	// We read the whole chart into memory, this should be okay since the chart
	// tarball needs to be small enough to fit into a GRPC call
	gzf, err := gzip.NewReader(reader)
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
	chartYamlFileName := fixedName + "/Chart.yaml"
	filenames := map[string]string{
		chart.ValuesKey:    valuesFileName,
		chart.ReadmeKey:    readmeFileName,
		chart.SchemaKey:    schemaFileName,
		chart.ChartYamlKey: chartYamlFileName,
	}

	files, err := ExtractFilesFromTarball(filenames, tarf)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		chart.ValuesKey:    files[chart.ValuesKey],
		chart.ReadmeKey:    files[chart.ReadmeKey],
		chart.SchemaKey:    files[chart.SchemaKey],
		chart.ChartYamlKey: files[chart.ChartYamlKey],
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
				_, err := io.Copy(&b, tarf)
				if err != nil {
					return ret, err
				}
				ret[id] = b.String()
				break
			}
		}
	}
	return ret, nil
}
