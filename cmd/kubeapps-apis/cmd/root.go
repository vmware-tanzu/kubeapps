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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
)

var (
	cfgFile   string
	serveOpts server.ServeOptions
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeapps-apis",
	Short: "A plugin-based gRPC and HTTP API server for interacting with Kubernetes packages",
	Long: `kubeapps-apis is a plugin-based API server for interacting with Kubernetes packages.

The api service serves both gRPC and HTTP requests for the configured APIs.`,

	Run: func(cmd *cobra.Command, args []string) {
		server.Serve(serveOpts)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubeapps-apis.yaml)")

	rootCmd.Flags().IntVar(&serveOpts.Port, "port", 50051, "The port on which to run this api server. Both gRPC and HTTP requests will be served on this port.")
	rootCmd.Flags().StringSliceVar(&serveOpts.PluginDirs, "plugin-dir", []string{"."}, "A directory to be scanned for .so plugins. May be specified multiple times.")

	rootCmd.Flags().StringVar(&serveOpts.ClustersConfigPath, "clusters-config-path", "", "Configuration for clusters")
	rootCmd.Flags().StringVar(&serveOpts.PinnipedProxyURL, "pinniped-proxy-url", "http://kubeapps-internal-pinniped-proxy.kubeapps:3333", "internal url to be used for requests to clusters configured for credential proxying via pinniped")
	rootCmd.Flags().BoolVar(&serveOpts.UnsafeUseDemoSA, "unsafe-use-demo-sa", false, "if true, it will create and use a privileged Service Account for interacting with the resources instead of acting on a user's behalf.")
	rootCmd.Flags().BoolVar(&serveOpts.UnsafeLocalDevKubeconfig, "unsafe-local-dev-kubeconfig", false, "if true, it will use the local kubeconfig at the KUBECONFIG env var instead of using the inCluster configuration.")
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

		// Search config in home directory with name ".kubeapps-apis" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kubeapps-apis")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
