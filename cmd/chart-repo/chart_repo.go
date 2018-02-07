/*
Copyright (c) 2017-2018 Bitnami

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

var RootCmd = &cobra.Command{
	Use:   "chart-repo",
	Short: "Kubeapps Chart Repository utility",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	cmd := RootCmd
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cmds := []*cobra.Command{syncCmd, deleteCmd}

	for _, cmd := range cmds {
		RootCmd.AddCommand(cmd)
		cmd.Flags().String("mongo-url", "localhost", "MongoDB URL (see https://godoc.org/labix.org/v2/mgo#Dial for format)")
		cmd.Flags().String("mongo-database", "charts", "MongoDB database")
		cmd.Flags().String("mongo-user", "", "MongoDB user")
		cmd.Flags().Bool("debug", false, "verbose logging")
	}
}
