package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
)

var upCmd = &cobra.Command{
	Use:   "up FLAG",
	Short: "install KubeApps components",
	Long:  `install KubeApps components`,
	Run: func(cmd *cobra.Command, args []string) {
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logrus.Fatal(err.Error())
		}

		out, err := cmd.Flags().GetString("out")
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
	upCmd.Flags().StringP("namespace", "", api.NamespaceDefault, "Specify namespace for the KubeApps components")
	upCmd.Flags().Bool("dry-run", false, "Provides output to be submitted to the server")
	upCmd.Flags().StringP("out", "o", "", "Output format. One of: json|yaml")
}
