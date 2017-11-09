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
	"fmt"
	"os"
	"strings"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/kubeapps/kubeapps/pkg/gke"
	"github.com/ksonnet/kubecfg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
)

const (
	GcTag = "bitnami/kubeapps"
)

var upCmd = &cobra.Command{
	Use:   "up FLAG",
	Short: "Install KubeApps components.",
	Long: `Install KubeApps components.

List of components that kubeapps up installs:

- Kubeless (https://github.com/kubeless/kubeless)
- Sealed-Secrets (https://github.com/bitnami/sealed-secrets)
- Helm/Tiller (https://github.com/kubernetes/helm)
- Kubeapps Dashboard (https://github.com/kubeapps/dashboard)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.ApplyCmd{
			DefaultNamespace: "default",
		}
		var err error
		c.Create = true

		c.DryRun, err = cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("can't get --dry-run flag: %v", err)
		}

		c.GcTag = GcTag

		c.ClientPool, c.Discovery, err = restClientPool()
		if err != nil {
			return fmt.Errorf("can't get Kubernetes client: %v", err)
		}

		// validate k8s version
		version, err := utils.FetchVersion(c.Discovery)
		if err != nil {
			return fmt.Errorf("can't verify Kubernetes version: %v", err)
		}
		if version.Major <= 1 && version.Minor < 7 {
			fmt.Println("warning: Kubernetes with RBAC enabled (v1.7+) is required to run Kubeapps")
			os.Exit(0)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("can't get current directory: %v", err)
		}
		wd := metadata.AbsPath(cwd)

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return fmt.Errorf("can't read kubeapps manifest: %v", err)
		}

		objs, err := parseObjects(manifest)
		if err != nil {
			return fmt.Errorf("can't parse kubeapps manifest: %v", err)
		}

		// k8s on GKE
		if ok, err := isGKE(c.Discovery); err != nil {
			return err
		} else if ok {
			gcloudPath, err := gke.SdkConfigPath()
			if err != nil {
				return fmt.Errorf("can't get sdk config path: %v", err)
			}

			user, err := gke.GetActiveUser(gcloudPath)
			if err != nil {
				return fmt.Errorf("can't get active gke user: %v", err)
			}

			crb, err := gke.BuildCrbObject(user)
			if err != nil {
				return fmt.Errorf("can't assign cluster-admin permission to the current user: %v", err)
			}

			//(tuna): we force the deployment ordering here:
			// this clusterrolebinding will be created before others for granting the proper permission.
			// when the installation finishes, it will be gc'd immediately.
			c.SkipGc = true
			err = c.Run(crb, wd)
			if err != nil {
				return fmt.Errorf("can't assign cluster-admin permission to the current user: %v", err)
			}
			c.SkipGc = false
		}

		if err = c.Run(objs, wd); err != nil {
			return fmt.Errorf("can't install kubeapps components: %v", err)
		}
		fmt.Println("successfully installed kubeapps")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("dry-run", false, "Provides output to be submitted to the server.")
}

func isGKE(disco discovery.DiscoveryInterface) (bool, error) {
	sv, err := disco.ServerVersion()
	if err != nil {
		return false, err
	}
	if strings.Contains(sv.GitVersion, "gke") {
		return true, nil
	}

	return false, nil
}
