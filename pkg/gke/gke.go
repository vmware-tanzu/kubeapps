package gke

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/go-ini/ini"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BuildCrbObject builds the clusterrolebinding for granting
// cluster-admin permission to the active user.
func BuildCrbObject(user string) ([]*unstructured.Unstructured, error) {
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
	crbs := []*unstructured.Unstructured{}
	crbs = append(crbs, crb)
	return crbs, nil
}

// GetActiveUser returns user logging to gcloud
func GetActiveUser(gcloudPath string) (string, error) {
	activeConfigPath := filepath.Join(gcloudPath, "active_config")
	activeConfig, err := ioutil.ReadFile(activeConfigPath)
	if err != nil {
		return "", fmt.Errorf("can't read file active_config: %v", err)
	}

	configPath := filepath.Join(gcloudPath, "configurations")
	if _, err := os.Stat(filepath.Join(configPath, "config_"+string(activeConfig))); os.IsNotExist(err) {
		return "", fmt.Errorf("the config file for active_config doesn't exist: %v", err)
	}

	cfg, err := ini.Load(filepath.Join(configPath, "config_"+string(activeConfig)))
	if err != nil {
		return "", fmt.Errorf("can't load file for the active_config: %v", err)
	}

	core, err := cfg.GetSection("core")
	if err != nil {
		return "", fmt.Errorf("can't get section [core]: %v", err)
	}
	account, err := core.GetKey("account")
	if err != nil {
		return "", fmt.Errorf("can't get key account: %v", err)
	}

	return account.Value(), nil
}

// SdkConfigPath returns path to gcloud sdk config
var SdkConfigPath = func() (string, error) {
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
