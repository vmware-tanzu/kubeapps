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

	"github.com/kubeapps/kubeapps/cmd/assetsvc/server"
	"github.com/spf13/cobra"
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
	cobra.OnInitialize(initConfig)
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
