/*
Copyright 2021 VMware. All Rights Reserved.

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

	"github.com/kubeapps/kubeapps/cmd/asset-syncer/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "k8s.io/klog/v2"
)

var (
	cfgFile   string
	serveOpts server.Config
	// This Version var is updated during the build
	// see the -ldflags option in the cmd/asset-syncer/Dockerfile
	version = "devel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd *cobra.Command
var syncCmd *cobra.Command
var deleteCmd *cobra.Command
var invalidateCacheCmd *cobra.Command

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "asset-syncer",
		Short: "Asset Synchronization utility",
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Infof("asset-syncer has been configured with: %#v", serveOpts)
		},
		Version: "devel",
	}
}

// syncCmd represents the sync command
func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [REPO NAME] [REPO URL] [REPO TYPE]",
		Short: "Add a new chart repository, and resync its charts periodically",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Sync(serveOpts, version, args)
		},
		Version: "devel",
	}
}

// deleteCmd represents the delete command
func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [REPO NAME]",
		Short: "delete a package repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Delete(serveOpts, args)
		},
		Version: "devel",
	}
}

// invalidateCacheCmd represents the invalidate-cache command
func newInvalidateCacheCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "invalidate-cache",
		Short: "removes all data so the cache can be rebuilt",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.InvalidateCache(serveOpts, args)
		},
		Version: "devel",
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Create new commands
	rootCmd = newRootCmd()
	syncCmd = newSyncCmd()
	deleteCmd = newDeleteCmd()
	invalidateCacheCmd = newInvalidateCacheCmd()

	// Set the flags of each command
	setRootFlags(rootCmd)
	setSyncFlags(syncCmd)

	// Set version
	rootCmd.SetVersionTemplate(version)

	// Enrich the config object with environmental information
	serveOpts.DatabasePassword = os.Getenv("DB_PASSWORD")
	serveOpts.KubeappsNamespace = os.Getenv("POD_NAMESPACE")
	serveOpts.AuthorizationHeader = os.Getenv("AUTHORIZATION_HEADER")
	serveOpts.DockerConfigJson = os.Getenv("DOCKER_CONFIG_JSON")
	serveOpts.UserAgent = server.GetUserAgent(version, serveOpts.UserAgent)

	// Register each command to the root cmd
	cmds := []*cobra.Command{syncCmd, deleteCmd, invalidateCacheCmd}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
}

func setRootFlags(c *cobra.Command) {
	c.PersistentFlags().StringVar(&serveOpts.DatabaseURL, "database-url", "localhost:5432", "Database URL")
	c.PersistentFlags().StringVar(&serveOpts.DatabaseName, "database-name", "charts", "Name of the database to use")
	c.PersistentFlags().StringVar(&serveOpts.DatabaseUser, "database-user", "", "Database user")
	c.PersistentFlags().StringVar(&serveOpts.Namespace, "namespace", "", "Namespace of the repository being synced")
	c.PersistentFlags().StringVar(&serveOpts.UserAgent, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	c.PersistentFlags().BoolVar(&serveOpts.Debug, "debug", false, "verbose logging")
	c.PersistentFlags().BoolVar(&serveOpts.TlsInsecureSkipVerify, "tls-insecure-skip-verify", false, "Skip TLS verification")
	c.PersistentFlags().StringVar(&serveOpts.FilterRules, "filter-rules", "", "JSON blob with the rules to filter assets")
	c.PersistentFlags().BoolVar(&serveOpts.PassCredentials, "pass-credentials", false, "pass credentials to all domains")
	c.PersistentFlags().StringVar(&serveOpts.GlobalReposNamespace, "global-repos-namespace", "kubeapps", "Namespace for global repos")
}

func setSyncFlags(c *cobra.Command) {
	c.Flags().StringSliceVar(&serveOpts.OciRepositories, "oci-repositories", []string{}, "List of OCI Repositories in case the type is OCI")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kubeops" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kubeops")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
