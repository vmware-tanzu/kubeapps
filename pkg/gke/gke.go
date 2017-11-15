package gke

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func BuildCrbObject() (*unstructured.Unstructured, error) {
	user, err := getActiveUser()
	if err != nil {
		return nil, err
	}

	crb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "ClusterRoleBinding",
			"apiVersion": "rbac.authorization.k8s.io/v1beta1",
			"metadata": map[string]interface{}{
				"name": "kubeapps-cluster-admin",
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "ClusterRole",
				"name":     "cluster-admin",
			},
			"subjects": []map[string]interface{}{
				{
					"apiGroup": "rbac.authorization.k8s.io",
					"kind":     "User",
					"name":     user,
				},
			},
		},
	}
	return crb, nil
}

func getActiveUser() (string, error) {
	gcloudPath, err := sdkConfigPath()
	if err != nil {
		return "", err
	}
	activeConfigPath := filepath.Join(gcloudPath, "active_config")
	activeConfig, err := ioutil.ReadFile(activeConfigPath)
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(gcloudPath, "configurations")
	f, err := os.Open(filepath.Join(configPath, "config_"+string(activeConfig)))
	if err != nil {
		return "", err
	}
	defer f.Close()
	ini, err := parseINI(f)
	if err != nil {
		return "", err
	}
	core, ok := ini["core"]
	if !ok {
		return "", err
	}
	active, ok := core["account"]
	if !ok {
		return "", err
	}

	return active, nil
}

func parseINI(ini io.Reader) (map[string]map[string]string, error) {
	result := map[string]map[string]string{
		"": {}, // root section
	}
	scanner := bufio.NewScanner(ini)
	currentSection := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			result[currentSection] = map[string]string{}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] != "" {
			result[currentSection][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning ini: %v", err)
	}
	return result, nil
}

var sdkConfigPath = func() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud"), nil
	}
	homeDir := guessUnixHomeDir()
	if homeDir == "" {
		return "", fmt.Errorf("unable to get current user home directory: os/user lookup failed; $HOME is empty")
	}
	return filepath.Join(homeDir, ".config", "gcloud"), nil
}

func guessUnixHomeDir() string {
	// Prefer $HOME over user.Current due to glibc bug: golang.org/issue/13470
	if v := os.Getenv("HOME"); v != "" {
		return v
	}
	// Else, fall back to user.Current:
	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}
	return ""
}
