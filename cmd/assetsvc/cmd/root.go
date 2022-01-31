// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"flag"
	"os"

	"github.com/kubeapps/kubeapps/cmd/assetsvc/server"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	log "k8s.io/klog/v2"
)

var (
	cfgFile   string
	serveOpts server.ServeOptions
	// This Version var is updated during the build
	// see the -ldflags option in the cmd/kubeops/Dockerfile
	version = "devel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd *cobra.Command

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assetsvc",
		Short: "assetsvc is a micro-service that creates an API endpoint for accessing the Helm API and Kubernetes resources.",
		PreRun: func(cmd *cobra.Command, args []string) {
			log.Infof("assetsvc has been configured with: %#v", serveOpts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Serve(serveOpts)
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
	log.InitFlags(nil)
	cobra.OnInitialize(initConfig)
	//set initial value of verbosity
	err := flag.Set("v", "3")
	if err != nil {
		log.Errorf("Error parsing verbosity: %v", viper.ConfigFileUsed())
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	rootCmd = newRootCmd()
	rootCmd.SetVersionTemplate(version)
	setFlags(rootCmd)

	serveOpts.DbPassword = os.Getenv("DB_PASSWORD")
	serveOpts.KubeappsNamespace = os.Getenv("POD_NAMESPACE")
}

func setFlags(c *cobra.Command) {
	c.Flags().StringVar(&serveOpts.DbURL, "database-url", "localhost", "Database URL")
	c.Flags().StringVar(&serveOpts.DbUsername, "database-user", "root", "Database user")
	c.Flags().StringVar(&serveOpts.DbName, "database-name", "charts", "Database name")
	c.Flags().StringVar(&serveOpts.GlobalReposNamespace, "global-repos-namespace", "kubeapps", "Namespace for global repos")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kubeops" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kubeops")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Errorf("Using config file: %v", viper.ConfigFileUsed())
	}
}
