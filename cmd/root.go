package cmd

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/kubecfg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Adding explicitely the GCP auth plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	// VERSION will be overwritten automatically by the build system
	VERSION = "devel"
)

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:   "kubeapps",
	Short: "Kubeapps Installer manages to install Kubeapps components to your cluster",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStderr()
		logrus.SetOutput(out)
		return nil
	},
}

func parseObjects(manifest string) ([]*unstructured.Unstructured, error) {
	r := strings.NewReader(manifest)
	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	ret := []runtime.Object{}
	for {
		bytes, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			continue
		}
		jsondata, err := yaml.ToJSON(bytes)
		if err != nil {
			return nil, err
		}
		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return utils.FlattenToV1(ret), nil
}

func restClientPool() (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	conf, err := buildOutOfClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	discoCache := utils.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	pool := dynamic.NewClientPool(conf, mapper, pathresolver)
	return pool, discoCache, nil
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := getHome()
		if err != nil {
			return nil, err
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func getHome() (string, error) {
	home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		for _, h := range []string{"HOME", "USERPROFILE"} {
			if home = os.Getenv(h); home != "" {
				return home, nil
			}
		}
	} else {
		return home, nil
	}

	return "", errors.New("can't get home directory")
}
