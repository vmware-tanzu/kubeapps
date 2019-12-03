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
	databaseType     string
	databaseURL      string
	databaseName     string
	databaseUser     string
	databasePassword string
	debug            bool
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
	rootCmd.PersistentFlags().StringVar(&databaseType, "database-type", "mongodb", "Database to use. Choice: mongodb, postgresql")
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database-url", "localhost", "MongoDB URL (see https://godoc.org/github.com/globalsign/mgo#Dial for format)")
	rootCmd.PersistentFlags().StringVar(&databaseName, "database-name", "charts", "MongoDB database")
	rootCmd.PersistentFlags().StringVar(&databaseUser, "database-user", "", "MongoDB user")
	// see version.go
	rootCmd.PersistentFlags().StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "verbose logging")

	databasePassword = os.Getenv("DB_PASSWORD")

	cmds := []*cobra.Command{syncCmd, deleteCmd}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(versionCmd)
}
