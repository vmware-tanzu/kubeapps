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
	cmds := []*cobra.Command{syncCmd, deleteCmd}

	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
		cmd.Flags().String("mongo-url", "localhost", "MongoDB URL (see https://godoc.org/github.com/globalsign/mgo#Dial for format)")
		cmd.Flags().String("mongo-database", "charts", "MongoDB database")
		cmd.Flags().String("mongo-user", "", "MongoDB user")
		// see version.go
		cmd.Flags().StringVarP(&userAgentComment, "user-agent-comment", "", "", "UserAgent comment used during outbound requests")
		cmd.Flags().Bool("debug", false, "verbose logging")
	}
	rootCmd.AddCommand(versionCmd)
}
