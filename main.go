//go:generate ./statik -f -src=./static -dest=./generated
package main

import (
	"os"

	"github.com/kubeapps/installer/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}
