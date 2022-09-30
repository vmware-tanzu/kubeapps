// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"flag"
	"strings"

	
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/server"
	corev1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"
)

var (
	cfgFile   string
	serveOpts server.Config
	// This Version var is updated during the build
	// see the -ldflags option in the cmd/apprepository-controller/Dockerfile
	version = "devel"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd *cobra.Command

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apprepository-controller",
		Short: "Apprepository-controller is a Kubernetes controller for managing package repositories added to Kubeapps.",
		PreRun: func(cmd *cobra.Command, args []string) {
			initServerOpts()
			log.Infof("apprepository-controller has been configured with: %#v", serveOpts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return server.Serve(serveOpts)
		},
		Version: "devel",
	}
}

// initServerOpts initialises the server options which are dependent
// on the parsed arguments.
func initServerOpts() {
	serveOpts.ImagePullSecretsRefs = getImagePullSecretsRefs(serveOpts.RepoSyncImagePullSecrets)
	serveOpts.ParsedCustomAnnotations = parseLabelsAnnotations(serveOpts.CustomAnnotations)
	serveOpts.ParsedCustomLabels = parseLabelsAnnotations(serveOpts.CustomLabels)
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
	err := flag.Set("v", "3")
	if err != nil {
		log.Errorf("Error parsing verbosity: %v", viper.ConfigFileUsed())
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func setFlags(c *cobra.Command) {
	c.Flags().StringVar(&serveOpts.Kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	c.Flags().StringVar(&serveOpts.APIServerURL, "apiserver", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	c.Flags().StringVar(&serveOpts.RepoSyncImage, "repo-sync-image", "docker.io/kubeapps/asset-syncer:latest", "container repo/image to use in CronJobs")
	c.Flags().StringSliceVar(&serveOpts.RepoSyncImagePullSecrets, "repo-sync-image-pullsecrets", nil, "optional reference to secrets in the same namespace to use for pulling the image used by this pod")
	c.Flags().StringVar(&serveOpts.RepoSyncCommand, "repo-sync-cmd", "/chart-repo", "command used to sync/delete repos for repo-sync-image")
	c.Flags().StringVar(&serveOpts.KubeappsNamespace, "namespace", "kubeapps", "Namespace to discover AppRepository resources")
	c.Flags().StringVar(&serveOpts.GlobalPackagingNamespace, "global-repos-namespace", "kubeapps", "Namespace for global repos")
	c.Flags().BoolVar(&serveOpts.ReposPerNamespace, "repos-per-namespace", true, "Defaults to watch for repos in all namespaces. Switch to false to watch only the configured namespace.")
	c.Flags().StringVar(&serveOpts.DBURL, "database-url", "localhost", "Database URL")
	c.Flags().StringVar(&serveOpts.DBUser, "database-user", "root", "Database user")
	c.Flags().StringVar(&serveOpts.DBName, "database-name", "charts", "Database name")
	c.Flags().StringVar(&serveOpts.DBSecretName, "database-secret-name", "kubeapps-db", "Kubernetes secret name for database credentials")
	c.Flags().StringVar(&serveOpts.DBSecretKey, "database-secret-key", "postgresql-root-password", "Kubernetes secret key used for database credentials")
	c.Flags().StringVar(&serveOpts.UserAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	c.Flags().StringVar(&serveOpts.Crontab, "crontab", "*/10 * * * *", "CronTab to specify schedule")
	// TTLSecondsAfterFinished specifies the number of seconds a sync job should live after finishing.
	// The support for this is currently beta in K8s (v1.21), older versions require a feature gate being set to enable it.
	// See https://kubernetes.io/docs/concepts/workloads/controllers/job/#clean-up-finished-jobs-automatically
	c.Flags().StringVar(&serveOpts.TTLSecondsAfterFinished, "ttl-lifetime-afterfinished-job", "3600", "Lifetime limit after which the resource Jobs are deleted expressed in seconds by default is 3600 (1h) ")
	c.Flags().StringSliceVar(&serveOpts.CustomAnnotations, "custom-annotations", []string{""}, "optional annotations to be passed to the generated CronJobs, Jobs and Pods objects. For example: my/annotation=foo")
	c.Flags().StringSliceVar(&serveOpts.CustomLabels, "custom-labels", []string{""}, "optional labels to be passed to the generated CronJobs, Jobs and Pods objects. For example: my/label=foo")
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

// getImagePullSecretsRefs gets the []string of Secrets names from the
// StringSliceVar flag list passed in the repoSyncImagePullSecrets arg
func getImagePullSecretsRefs(imagePullSecretsRefsArr []string) []corev1.LocalObjectReference {
	var imagePullSecretsRefs []corev1.LocalObjectReference

	// getting and appending a []LocalObjectReference for each ImagePullSecret passed
	for _, imagePullSecretName := range imagePullSecretsRefsArr {
		imagePullSecretsRefs = append(imagePullSecretsRefs, corev1.LocalObjectReference{Name: imagePullSecretName})
	}
	return imagePullSecretsRefs
}

// parseLabelsAnnotations transform an array of string "foo=bar" into a map["foo"]="bar"
func parseLabelsAnnotations(textArr []string) map[string]string {
	textMap := map[string]string{}
	for _, text := range textArr {
		if text != "" {
			parts := strings.Split(text, "=")
			if len(parts) != 2 {
				log.Errorf("Cannot parse '%s'", text)
			}
			textMap[parts[0]] = parts[1]
		}
	}
	return textMap
}
