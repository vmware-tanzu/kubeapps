// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
)

const (
	flagGracePeriod = "grace-period"
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	addEnvCmdFlags(deleteCmd)
	bindClientGoFlags(deleteCmd)
	bindJsonnetFlags(deleteCmd)
	deleteCmd.PersistentFlags().Int64(flagGracePeriod, -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored")
}

var deleteCmd = &cobra.Command{
	Use:   "delete [env-name] [-f <file-or-dir>]",
	Short: "Delete Kubernetes resources described in local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'delete' takes at most a single argument, that is the name of the environment")
		}

		flags := cmd.Flags()
		var err error

		c := kubecfg.DeleteCmd{}

		c.GracePeriod, err = flags.GetInt64(flagGracePeriod)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		c.ClientPool, c.Discovery, err = restClientPool(cmd, envSpec.env)
		if err != nil {
			return err
		}

		c.DefaultNamespace, err = defaultNamespace()
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs)
	},
	Long: `Delete Kubernetes resources from a cluster, as described in the local
configuration.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.`,
	Example: `  # Delete all resources described in a ksonnet application, from the 'dev'
  # environment. Can be used in any subdirectory of the application.
  ksonnet delete dev

  # Delete resources described in a YAML file. Automatically picks up the
  # cluster's location from '$KUBECONFIG'.
  ksonnet delete -f ./pod.yaml

  # Delete resources described in the JSON file from the 'dev' environment. Can
  # be used in any subdirectory of the application.
  ksonnet delete dev -f ./pod.json

  # Delete resources described in a YAML file, and running in the cluster
  # specified by the current context in specified kubeconfig file.
  ksonnet delete --kubeconfig=./kubeconfig -f ./pod.yaml`,
}
