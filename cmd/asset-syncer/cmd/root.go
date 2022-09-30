// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"flag"
	"os"

	
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vmware-tanzu/kubeapps/cmd/asset-syncer/server"
	log "k8s.io/klog/v2"
)

var (
	cfgFile   string
	serveOpts server.Config
	// This Version var is updated during the build
	// see the -ldflags option in the cmd/asset-syncer/Dockerfile
	version = "devel"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "asset-syncer",
		Short: "Asset Synchronization utility",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			serveOpts.UserAgent = server.GetUserAgent(version, serveOpts.UserAgentComment)
			serveOptsCopy := serveOpts
			serveOptsCopy.DatabasePassword = "REDACTED"
			log.Infof("asset-syncer has been configured with: %#v", serveOptsCopy)
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

func newCmd() *cobra.Command {
	// Create new commands
	c := newRootCmd()
	syncCmd := newSyncCmd()
	deleteCmd := newDeleteCmd()
	invalidateCacheCmd := newInvalidateCacheCmd()

	// Set the flags of each command
	setRootFlags(c)
	setSyncFlags(syncCmd)

	// Set version
	c.SetVersionTemplate(version)

	// Register each command to the root cmd
	cmds := []*cobra.Command{syncCmd, deleteCmd, invalidateCacheCmd}
	for _, cmd := range cmds {
		c.AddCommand(cmd)
	}

	return c
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(newCmd().Execute())
}

func init() {
	log.InitFlags(nil)
	cobra.OnInitialize(initConfig)
	//set initial value of verbosity
	err := flag.Set("v", "3")
	if err != nil {
		log.Errorf("Error parsing verbosity: %v", viper.ConfigFileUsed())
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Enrich the config object with environmental information
	serveOpts.DatabasePassword = os.Getenv("DB_PASSWORD")
	serveOpts.KubeappsNamespace = os.Getenv("POD_NAMESPACE")
	serveOpts.AuthorizationHeader = os.Getenv("AUTHORIZATION_HEADER")
	serveOpts.DockerConfigJson = os.Getenv("DOCKER_CONFIG_JSON")
}

func setRootFlags(c *cobra.Command) {
	c.PersistentFlags().StringVar(&serveOpts.DatabaseURL, "database-url", "localhost:5432", "Database URL")
	c.PersistentFlags().StringVar(&serveOpts.DatabaseName, "database-name", "charts", "Name of the database to use")
	c.PersistentFlags().StringVar(&serveOpts.DatabaseUser, "database-user", "", "Database user")
	c.PersistentFlags().StringVar(&serveOpts.Namespace, "namespace", "", "Namespace of the repository being synced")
	c.PersistentFlags().StringVar(&serveOpts.UserAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	c.PersistentFlags().BoolVar(&serveOpts.Debug, "debug", false, "verbose logging")
	c.PersistentFlags().BoolVar(&serveOpts.TlsInsecureSkipVerify, "tls-insecure-skip-verify", false, "Skip TLS verification")
	c.PersistentFlags().StringVar(&serveOpts.FilterRules, "filter-rules", "", "JSON blob with the rules to filter assets")
	c.PersistentFlags().BoolVar(&serveOpts.PassCredentials, "pass-credentials", false, "pass credentials to all domains")
	c.PersistentFlags().StringVar(&serveOpts.GlobalPackagingNamespace, "global-repos-namespace", "kubeapps", "Namespace for global repos")
}

func setSyncFlags(c *cobra.Command) {
	c.Flags().StringSliceVar(&serveOpts.OciRepositories, "oci-repositories", []string{}, "List of OCI Repositories in case the type is OCI")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Errorf("Using config file: %v", viper.ConfigFileUsed())
	}
}
