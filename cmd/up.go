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

		kaFile, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		if kaFile == "" {
			kaFile, err = cmd.Flags().GetString("path")
			if err != nil {
				return err
			}
			if kaFile == "" {
				home, err := getHome()
				if err != nil {
					return err
				}
				kaFile = filepath.Join(home, KUBEAPPS_DIR)
			}
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

		objs, err := parseObjects(kaFile)
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
