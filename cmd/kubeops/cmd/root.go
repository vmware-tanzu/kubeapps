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

	"github.com/kubeapps/kubeapps/cmd/kubeops/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	serveOpts server.ServeOptions
	// This Version var is updated during the build
	// see the -ldflags option in the cmd/kubeops/Dockerfile
	version = "devzzzel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeops",
	Short: "Kubeops is micro-service that creates an API endpoint for accessing the Helm API and Kubernetes resources.",
	Long:  "Kubeops is micro-service that creates an API endpoint for accessing the Helm API and Kubernetes resources.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Serve(serveOpts)
	},
	Version: "devel",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetVersionTemplate(version)

	rootCmd.Flags().StringVar(&serveOpts.AssetsvcURL, "assetsvc-url", "https://kubeapps-internal-assetsvc:8080", "URL to the internal assetsvc")
	rootCmd.Flags().StringVar(&serveOpts.HelmDriverArg, "helm-driver", "", "which Helm driver type to use")
	rootCmd.Flags().IntVar(&serveOpts.ListLimit, "list-max", 256, "maximum number of releases to fetch")
	rootCmd.Flags().StringVar(&serveOpts.UserAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	rootCmd.Flags().Int64Var(&serveOpts.Timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
	rootCmd.Flags().StringVar(&serveOpts.ClustersConfigPath, "clusters-config-path", "", "Configuration for clusters")
	rootCmd.Flags().StringVar(&serveOpts.PinnipedProxyURL, "pinniped-proxy-url", "http://kubeapps-internal-pinniped-proxy.kubeapps:3333", "internal url to be used for requests to clusters configured for credential proxying via pinniped")
	rootCmd.Flags().IntVar(&serveOpts.Burst, "burst", 15, "internal burst capacity")
	rootCmd.Flags().Float32Var(&serveOpts.Qps, "qps", 10, "internal QPS rate")
	rootCmd.Flags().StringVar(&serveOpts.NamespaceHeaderName, "namespace-header-name", "", "name of the header field, e.g. namespace-header-name=X-Consumer-Groups")
	rootCmd.Flags().StringVar(&serveOpts.NamespaceHeaderPattern, "namespace-header-pattern", "", "regular expression that matches only single group, e.g. namespace-header-pattern=^namespace:([\\w]+):\\w+$, to match namespace:ns:read")

	serveOpts.UserAgent = getUserAgent(version, serveOpts.UserAgentComment)

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

// Returns the user agent to be used during calls to the chart repositories
// Examples:
// kubeops/devel
// kubeops/2.3.4 (kubeapps v2.3.4-beta4)
func getUserAgent(version, userAgentComment string) string {
	ua := "kubeops/" + version
	if userAgentComment != "" {
		ua = fmt.Sprintf("%s (%s)", ua, userAgentComment)
	}
	return ua
}
