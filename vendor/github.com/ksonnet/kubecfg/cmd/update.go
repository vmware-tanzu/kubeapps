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

func init() {
	RootCmd.AddCommand(updateCmd)

	addEnvCmdFlags(updateCmd)
	bindClientGoFlags(updateCmd)
	bindJsonnetFlags(updateCmd)
	updateCmd.PersistentFlags().Bool(flagCreate, true, "Create missing resources")
	updateCmd.PersistentFlags().Bool(flagSkipGc, false, "Don't perform garbage collection, even with --"+flagGcTag)
	updateCmd.PersistentFlags().String(flagGcTag, "", "Add this tag to updated objects, and garbage collect existing objects with this tag and not in config")
	updateCmd.PersistentFlags().Bool(flagDryRun, false, "Perform only read-only operations")
}

var updateCmd = &cobra.Command{
	Deprecated: "NOTE: Command 'update' is deprecated, use 'apply' instead",
	Hidden:     true,
	Use:        "update [<env>|-f <file-or-dir>]",
	Short: `[DEPRECATED] Update (or optionally create) Kubernetes resources on the cluster using the
local configuration. Accepts JSON, YAML, or Jsonnet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'update' takes at most a single argument, that is the name of the environment")
		}

		flags := cmd.Flags()
		var err error

		c := kubecfg.ApplyCmd{}

		c.Create, err = flags.GetBool(flagCreate)
		if err != nil {
			return err
		}

		c.GcTag, err = flags.GetString(flagGcTag)
		if err != nil {
			return err
		}

		c.SkipGc, err = flags.GetBool(flagSkipGc)
		if err != nil {
			return err
		}

		c.DryRun, err = flags.GetBool(flagDryRun)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		c.ClientPool, c.Discovery, err = restClientPool(cmd, nil)
		if err != nil {
			return err
		}

		c.DefaultNamespace, err = defaultNamespace()
		if err != nil {
			return err
		}

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
	Long: `NOTE: Command 'update' is deprecated, use 'apply' instead.

Update (or optionally create) Kubernetes resources on the cluster using the
local configuration. Use the '--create' flag to control whether we create them
if they do not exist (default: true).

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.`,
	Example: `  # Create or update all resources described in a ksonnet application, and
  # running in the 'dev' environment. Can be used in any subdirectory of the
  # application.
  ksonnet update dev

  # Create or update resources described in a YAML file. Automatically picks up
  # the cluster's location from '$KUBECONFIG'.
  ksonnet appy -f ./pod.yaml

  # Update resources described in a YAML file, and running in cluster referred
  # to by './kubeconfig'.
  ksonnet update --kubeconfig=./kubeconfig -f ./pod.yaml

  # Display set of actions we will execute when we run 'update'.
  ksonnet update dev --dry-run`,
}
