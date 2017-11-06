package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kubeapps Installer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Kubeapps Installer version: " + VERSION)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
