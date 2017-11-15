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
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/kubeapps/installer/pkg/gke"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var downCmd = &cobra.Command{
	Use:   "down FLAG",
	Short: "Uninstall KubeApps components.",
	Long:  `Uninstall KubeApps components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.DeleteCmd{
			DefaultNamespace: "default",
		}

		var err error
		c.GracePeriod, err = cmd.Flags().GetInt64("grace-period")
		if err != nil {
			return err
		}

		c.ClientPool, c.Discovery, err = restClientPool()
		if err != nil {
			return err
		}
		provider, err := cmd.Flags().GetString("cloud-provider")
		if err != nil {
			return err
		}

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return err
		}
		objs, err := parseObjects(manifest)
		if err != nil {
			return err
		}

		// k8s on GKE
		crb := &unstructured.Unstructured{}
		if provider == "gke" {
			crb, err = gke.BuildCrbObject()
			if err != nil {
				return err
			}
			objs = append(objs, crb)
		}

		return c.Run(objs)
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
	downCmd.Flags().Int64("grace-period", -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored.")
}
