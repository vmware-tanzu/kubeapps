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
	flagEnvName = "name"
	flagEnvURI  = "uri"
)

func init() {
	RootCmd.AddCommand(envCmd)
	bindClientGoFlags(envCmd)

	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envRmCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)

	// TODO: We need to make this default to checking the `kubeconfig` file.
	envAddCmd.PersistentFlags().String(flagAPISpec, "version:v1.7.0",
		"Manually specify API version from OpenAPI schema, cluster, or Kubernetes version")

	envSetCmd.PersistentFlags().String(flagEnvName, "",
		"Specify name to rename environment to. Name must not already exist")
	envSetCmd.PersistentFlags().String(flagEnvURI, "",
		"Specify URI to point environment cluster to a new location")
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: `Manage ksonnet environments`,
	Long: `An environment acts as a sort of "named cluster", allowing for commands like
'ksonnet apply dev', which applies the ksonnet application to the "dev cluster".
Additionally, environments allow users to cache data about the cluster it points
to, including data needed to run 'verify', and a version of ksonnet-lib that is
generated based on the flags the API server was started with (e.g., RBAC enabled
or not).

An environment contains no user-specific data (such as the private key
often contained in a kubeconfig file), and

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application. For example, in the example below, there are two
environments: 'default' and 'us-west/staging'. Each contains a cached version of
ksonnet-lib, and a 'spec.json' that contains the URI and server cert that
uniquely identifies the cluster.

environments/
  default/           [Default generated environment]
    k.libsonnet
    k8s.libsonnet
    swagger.json
    spec.json
  us-west/
    staging/         [Example of user-generated env]
      k.libsonnet
      k8s.libsonnet
      swagger.json
      spec.json      [This will contain the uri of the environment]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'env' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var envAddCmd = &cobra.Command{
	Use:   "add <env-name> <env-uri>",
	Short: "Add a new environment to a ksonnet project",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 2 {
			return fmt.Errorf("'env add' takes two arguments, the name and the uri of the environment, respectively")
		}

		envName := args[0]
		envURI := args[1]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		specFlag, err := flags.GetString(flagAPISpec)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvAddCmd(envName, envURI, specFlag, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},

	Long: `Add a new environment to a ksonnet project. Names are restricted to not
include punctuation, so names like '../foo' are not allowed.

An environment acts as a sort of "named cluster", allowing for commands like
'ksonnet apply dev', which applies the ksonnet application to the "dev cluster".
For more information on what an environment is and how they work, run 'help
env'.

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application, and hence 'env add' will add to this directory structure.
For example, in the example below, there are two environments: 'default' and
'us-west/staging'. 'env add' will add a similar directory to this environment.

environments/
  default/           [Default generated environment]
    k.libsonnet
    k8s.libsonnet
    swagger.json
    spec.json
  us-west/
    staging/         [Example of user-generated env]
      k.libsonnet
      k8s.libsonnet
      swagger.json
      spec.json      [This will contain the uri of the environment]`,
	Example: `  # Initialize a new staging environment at us-west. The directory
  # structure rooted at 'us-west' in the documentation above will be generated.
  ksonnet env add us-west/staging https://kubecfg-1.us-west.elb.amazonaws.com

  # Initialize a new staging environment at us-west, using the OpenAPI specification
  # generated in the Kubernetes v1.7.1 build to generate 'ksonnet-lib'.
  ksonnet env add us-west/staging https://kubecfg-1.us-west.elb.amazonaws.com --api-spec=version:v1.7.1

  # Initialize a new development environment locally. This will overwrite the
  # default 'default' directory structure generated by 'ksonnet-init'.
  ksonnet env add default localhost:8000`,
}

var envRmCmd = &cobra.Command{
	Use:   "rm <env-name>",
	Short: "Delete an environment from a ksonnet project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'env rm' takes a single argument, that is the name of the environment")
		}

		envName := args[0]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvRmCmd(envName, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Delete an environment from a ksonnet project. This is the same
as removing the <env-name> environment directory and all files contained. If the
project exists in a hierarchy (e.g., 'us-east/staging') and deleting the
environment results in an empty environments directory (e.g., if deleting
'us-east/staging' resulted in an empty 'us-east/' directory), then all empty
parent directories are subsequently deleted.`,
	Example: `  # Remove the directory 'us-west/staging' and all contents
  #	in the 'environments' directory. This will also remove the parent directory
  # 'us-west' if it is empty.
  ksonnet env rm us-west/staging`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments in a ksonnet project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("'env list' takes zero arguments")
		}

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvListCmd(manager)
		if err != nil {
			return err
		}

		return c.Run(cmd.OutOrStdout())
	}, Long: `List all environments in a ksonnet project. This will
display the name and the URI of each environment within the ksonnet project.`,
}

var envSetCmd = &cobra.Command{
	Use:   "set <env-name> [parameter-flags]",
	Short: "Set environment fields such as the name, and cluster URI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 1 {
			return fmt.Errorf("'env set' takes a single argument, that is the name of the environment")
		}

		envName := args[0]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		desiredEnvName, err := flags.GetString(flagEnvName)
		if err != nil {
			return err
		}

		desiredEnvURI, err := flags.GetString(flagEnvURI)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvSetCmd(envName, desiredEnvName, desiredEnvURI, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Set environment fields such as the name, and cluster URI. Changing
the name of an environment will also update the directory structure in
'environments'.`,
	Example: `  # Updates the URI of the environment 'us-west/staging'.
  ksonnet env set us-west/staging --uri=http://example.com

  # Updates both the name and the URI of the environment 'us-west/staging'.
  # Updating the name will update the directory structure in 'environments'
  ksonnet env set us-west/staging --uri=http://example.com --name=us-east/staging`,
}
