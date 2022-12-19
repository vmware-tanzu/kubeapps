// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package tarutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"regexp"
	"strings"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	chart "github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
)

// Fetches helm chart details from a gzipped tarball
//
// name is expected in format "foo/bar" or "foo%2Fbar" if url-escaped
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

	_, fixedName, err := pkgutils.SplitPackageIdentifier(name)
	if err != nil {
		return nil, err
	}

	return FetchChartDetailFromTarball(reader, fixedName)
}

// Fetches helm chart details from a gzipped tarball
//
// name is expected in format "foo/bar" or "foo%2Fbar" if url-escaped
func FetchChartDetailFromTarball(reader io.Reader, tarballRootDir string) (map[string]string, error) {
	// We read the whole chart into memory, this should be okay since the chart
	// tarball needs to be small enough to fit into a GRPC call
	gzf, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzf.Close()

	tarf := tar.NewReader(gzf)

	prefix := ""
	if tarballRootDir != "" {
		prefix = tarballRootDir + "/"
	}

	readmeFileName := prefix + "README.md"
	valuesFileName := prefix + "values.yaml"
	schemaFileName := prefix + "values.schema.json"
	chartYamlFileName := prefix + "Chart.yaml"
	filenames := map[string]string{
		chart.DefaultValuesKey: valuesFileName,
		chart.ReadmeKey:        readmeFileName,
		chart.SchemaKey:        schemaFileName,
		chart.ChartYamlKey:     chartYamlFileName,
	}

	// Optionally search for files matching a regular expression, using the
	// template to provide the key.
	regexes := map[string]*regexp.Regexp{
		chart.DefaultValuesKey + "-$valuesType": regexp.MustCompile(prefix + `values-(?P<valuesType>\w+)\.yaml`),
	}

	return ExtractFilesFromTarball(filenames, regexes, tarf)
}

// ExtractFilesFromTarball returns the content of extracted files in a map.
//
// Files can be extracted by exact matches on the filename, or by regular
// expression matches. For exact matches, the key used in the resulting map
// is simply the key of the filename. For regex matches, a regexp template
// defines the key so that it can be expanded from the match.
func ExtractFilesFromTarball(filenames map[string]string, regexes map[string]*regexp.Regexp, tarf *tar.Reader) (map[string]string, error) {
	ret := make(map[string]string)
	for {
		header, err := tarf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ret, err
		}

		foundFile := false
		for id, f := range filenames {
			if strings.EqualFold(header.Name, f) {
				if s, err := readTarFileContent(tarf); err != nil {
					return ret, err
				} else {
					ret[id] = s
				}
				foundFile = true
				break
			}
		}
		if foundFile {
			continue
		}

		for template, pattern := range regexes {
			match := pattern.FindSubmatchIndex([]byte(header.Name))
			if match != nil {
				result := []byte{}
				result = pattern.ExpandString(result, template, header.Name, match)
				if s, err := readTarFileContent(tarf); err != nil {
					return ret, err
				} else {
					ret[string(result)] = s
				}
			}
		}
	}
	return ret, nil
}

func readTarFileContent(tarf *tar.Reader) (string, error) {
	var b bytes.Buffer
	_, err := io.Copy(&b, tarf)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
