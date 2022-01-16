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
	goflag "flag"
	"fmt"
	"os"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	log "k8s.io/klog/v2"
)

var (
	cfgFile   string
	serveOpts core.ServeOptions
	// This version var is updated during the build
	// see the -ldflags option in the Dockerfile
	version = "devel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd *cobra.Command

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kubeapps-apis",
		Short: "A plugin-based gRPC and HTTP API server for interacting with Kubernetes packages",
		Long: `kubeapps-apis is a plugin-based API server for interacting with Kubernetes packages.

The api service serves both gRPC and HTTP requests for the configured APIs.`,

		PreRun: func(cmd *cobra.Command, args []string) {
			log.Infof("kubeapps-apis has been configured with: %#v", serveOpts)
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
	rootCmd = newRootCmd()
	rootCmd.SetVersionTemplate(version)
	setFlags(rootCmd)
	//set initial value of verbosity
	goflag.Set("v", "3")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func setFlags(c *cobra.Command) {
	c.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubeapps-apis.yaml)")
	c.Flags().IntVar(&serveOpts.Port, "port", 50051, "The port on which to run this api server. Both gRPC and HTTP requests will be served on this port.")
	c.Flags().StringSliceVar(&serveOpts.PluginDirs, "plugin-dir", []string{"."}, "A directory to be scanned for .so plugins. May be specified multiple times.")
	c.Flags().StringVar(&serveOpts.ClustersConfigPath, "clusters-config-path", "", "Configuration for clusters")
	c.Flags().StringVar(&serveOpts.PluginConfigPath, "plugin-config-path", "", "Configuration for plugins")
	c.Flags().StringVar(&serveOpts.PinnipedProxyURL, "pinniped-proxy-url", "http://kubeapps-internal-pinniped-proxy.kubeapps:3333", "internal url to be used for requests to clusters configured for credential proxying via pinniped")
	c.Flags().StringVar(&serveOpts.GlobalReposNamespace, "global-repos-namespace", "kubeapps", "Namespace of global repositories")
	c.Flags().BoolVar(&serveOpts.UnsafeLocalDevKubeconfig, "unsafe-local-dev-kubeconfig", false, "if true, it will use the local kubeconfig at the KUBECONFIG env var instead of using the inCluster configuration.")
	c.Flags().Float32Var(&serveOpts.QPS, "kube-api-qps", 10.0, "set Kubernetes API client QPS limit")
	c.Flags().IntVar(&serveOpts.Burst, "kube-api-burst", 15, "set Kubernetes API client Burst limit")
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
