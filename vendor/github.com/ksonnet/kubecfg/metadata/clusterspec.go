package metadata

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	k8sVersionURLTemplate = "https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json"
)

func parseClusterSpec(specFlag string, fs afero.Fs) (ClusterSpec, error) {
	split := strings.SplitN(specFlag, ":", 2)
	if len(split) <= 1 || split[1] == "" {
		return nil, fmt.Errorf("Invalid API specification '%s'", specFlag)
	}

	switch split[0] {
	case "version":
		return &clusterSpecVersion{k8sVersion: split[1]}, nil
	case "file":
		abs, err := filepath.Abs(split[1])
		if err != nil {
			return nil, err
		}
		absPath := AbsPath(abs)
		return &clusterSpecFile{specPath: absPath, fs: fs}, nil
	case "url":
		return &clusterSpecLive{apiServerURL: split[1]}, nil
	default:
		return nil, fmt.Errorf("Could not parse cluster spec '%s'", specFlag)
	}
}

type clusterSpecFile struct {
	specPath AbsPath
	fs       afero.Fs
}

func (cs *clusterSpecFile) data() ([]byte, error) {
	return afero.ReadFile(cs.fs, string(cs.specPath))
}

func (cs *clusterSpecFile) resource() string {
	return string(cs.specPath)
}

type clusterSpecLive struct {
	apiServerURL string
}

func (cs *clusterSpecLive) data() ([]byte, error) {
	return nil, fmt.Errorf("Initializing from OpenAPI spec in live cluster is not implemented")
}

func (cs *clusterSpecLive) resource() string {
	return string(cs.apiServerURL)
}

type clusterSpecVersion struct {
	k8sVersion string
}

func (cs *clusterSpecVersion) data() ([]byte, error) {
	versionURL := fmt.Sprintf(k8sVersionURLTemplate, cs.k8sVersion)
	resp, err := http.Get(versionURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Recieved status code '%d' when trying to retrieve OpenAPI schema for cluster version '%s' from URL '%s'",
			resp.StatusCode, cs.k8sVersion, versionURL)
	}

	return ioutil.ReadAll(resp.Body)
}

func (cs *clusterSpecVersion) resource() string {
	return string(cs.k8sVersion)
}
