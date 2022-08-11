// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeops/server"
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
		Use:   "kubeops",
		Short: "Kubeops is a micro-service that creates an API endpoint for accessing the Helm API and Kubernetes resources.",
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
}

func setFlags(c *cobra.Command) {
	c.Flags().StringVar(&serveOpts.ClustersConfigPath, "clusters-config-path", "", "Configuration for clusters")
	c.Flags().StringVar(&serveOpts.PinnipedProxyURL, "pinniped-proxy-url", "http://kubeapps-internal-pinniped-proxy.kubeapps:3333", "internal url to be used for requests to clusters configured for credential proxying via pinniped")
	c.Flags().StringVar(&serveOpts.PinnipedProxyCACert, "pinniped-proxy-ca-cert", "", "Path to certificate authority to use with requests to pinniped-proxy service")
	c.Flags().IntVar(&serveOpts.Burst, "burst", 15, "internal burst capacity")
	c.Flags().Float32Var(&serveOpts.Qps, "qps", 10, "internal QPS rate")
	c.Flags().StringVar(&serveOpts.NamespaceHeaderName, "namespace-header-name", "", "name of the header field, e.g. namespace-header-name=X-Consumer-Groups")
	c.Flags().StringVar(&serveOpts.NamespaceHeaderPattern, "namespace-header-pattern", "", "regular expression that matches only single group, e.g. namespace-header-pattern=^namespace:([\\w]+):\\w+$, to match namespace:ns:read")

	// TODO(agamez): remove once a new version of the chart is released
	var deprecated string
	c.Flags().StringVar(&deprecated, "user-agent-comment", "", "(deprecated) UserAgent comment used during outbound requests")
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
