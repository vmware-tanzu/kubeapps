package cmd

import (
	"os"
	"path/filepath"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/spf13/cobra"
)

const (
	GcTag = "bitnami/kubeapps"
)

var upCmd = &cobra.Command{
	Use:   "up FLAG",
	Short: "install KubeApps components",
	Long:  `install KubeApps components`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.ApplyCmd{}
		ns, err := cmd.Flags().GetString("namespace")
		c.DefaultNamespace = ns
		if err != nil {
			return err
		}
		kaManifestDir, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		if kaManifestDir == "" {
			home, err := getHome()
			if err != nil {
				return err
			}
			kaManifestDir = filepath.Join(home, KUBEAPPS_DIR)
		}

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

		objs, err := parseObjects(kaManifestDir)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("dry-run", false, "Provides output to be submitted to the server")
	bindFlags(upCmd)
}
