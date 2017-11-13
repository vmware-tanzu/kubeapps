package cmd

import (
	"os"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/spf13/cobra"
)

const (
	GcTag = "bitnami/kubeapps"
)

var upCmd = &cobra.Command{
	Use:   "up FLAG",
	Short: "Install KubeApps components.",
	Long: `Install KubeApps components.

List of components that kubeapps up installs:

- Kubeless (https://github.com/kubeless/kubeless)
- Sealed-Secrets (https://github.com/bitnami/sealed-secrets)
- Helm/Tiller (https://github.com/kubernetes/helm)
- Kubeapps Dashboard (https://github.com/kubeapps/dashboard)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.ApplyCmd{
			DefaultNamespace: "default",
		}
		var err error
		c.Create = true

		c.DryRun, err = cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		c.GcTag = GcTag

		c.ClientPool, c.Discovery, err = restClientPool()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return err
		}
		objs, err := parseObjects(manifest)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("dry-run", false, "Provides output to be submitted to the server")
}
