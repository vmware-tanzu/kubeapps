/*
Copyright (c) 2018 The Helm Authors

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

package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	databaseURL           string
	databaseName          string
	databaseUser          string
	databasePassword      string
	debug                 bool
	namespace             string
	ociRepositories       []string
	tlsInsecureSkipVerify bool
	filterRules           string
	passCredentials       bool
)

var rootCmd = &cobra.Command{
	Use:   "asset-syncer",
	Short: "Asset Synchronization utility",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	cmd := rootCmd
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database-url", "localhost:5432", "Database URL")
	rootCmd.PersistentFlags().StringVar(&databaseName, "database-name", "charts", "Name of the database to use")
	rootCmd.PersistentFlags().StringVar(&databaseUser, "database-user", "", "Database user")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "Namespace of the repository being synced")
	// User agent configuration can be found in version.go. Check that file for more details
	rootCmd.PersistentFlags().StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVar(&tlsInsecureSkipVerify, "tls-insecure-skip-verify", false, "Skip TLS verification")
	rootCmd.PersistentFlags().StringVar(&filterRules, "filter-rules", "", "JSON blob with the rules to filter assets")
	rootCmd.PersistentFlags().BoolVar(&passCredentials, "pass-credentials", false, "pass credentials to all domains")

	databasePassword = os.Getenv("DB_PASSWORD")

	syncCmd.Flags().StringSliceVar(&ociRepositories, "oci-repositories", []string{}, "List of OCI Repositories in case the type is OCI")
	cmds := []*cobra.Command{syncCmd, deleteCmd, invalidateCacheCmd}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(versionCmd)
}
