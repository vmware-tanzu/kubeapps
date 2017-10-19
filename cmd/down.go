package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
)

var downCmd = &cobra.Command{
	Use:   "down FLAG",
	Short: "uninstall KubeApps components",
	Long:  `uninstall KubeApps components`,
	Run: func(cmd *cobra.Command, args []string) {
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			logrus.Fatal(err.Error())
		}
	},
}

func init() {
	downCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the KubeApps components")
	downCmd.Flags().Bool("dry-run", false, "Provides output to be submitted to the server")
}
