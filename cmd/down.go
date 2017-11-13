package cmd

import (
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down FLAG",
	Short: "Uninstall KubeApps components.",
	Long:  `Uninstall KubeApps components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.DeleteCmd{
			DefaultNamespace: "default",
		}

		var err error
		c.GracePeriod, err = cmd.Flags().GetInt64("grace-period")
		if err != nil {
			return err
		}

		c.ClientPool, c.Discovery, err = restClientPool()
		if err != nil {
			return err
		}

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return err
		}
		objs, err := parseObjects(manifest)
		if err != nil {
			return err
		}
		return c.Run(objs)
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
	downCmd.Flags().Int64("grace-period", -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored.")
}
