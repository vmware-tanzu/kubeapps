package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ksonnet/kubecfg/template"
	"github.com/ksonnet/kubecfg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const KUBEAPPS_DIR = ".kubeapps"

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:   "kubeapps",
	Short: "Manage KubeApps infrastructure",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStderr()
		logrus.SetOutput(out)
		return nil
	},
}

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the KubeApps components")
	cmd.Flags().String("path", "", "Specify folder contains the manifests")
	cmd.Flags().String("file", "", "Specify the kubeapps.jsonnet file")

}

func walkYaml(f *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".yaml" {
				*f = append(*f, path)
			}
		}

		return nil
	}
}

func parseObjects(path string) ([]*unstructured.Unstructured, error) {
	files := &[]string{}
	if filepath.Ext(path) == ".jsonnet" {
		files = &[]string{path}
	} else {
		err := filepath.Walk(path, walkYaml(files))
		if err != nil {
			return nil, err
		}
	}

	expander := &template.Expander{
		EnvJPath:   filepath.SplitList(os.Getenv("KUBECFG_JPATH")),
		FailAction: "warn",
		Resolver:   "noop",
	}
	return expander.Expand(*files)
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
