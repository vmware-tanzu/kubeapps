package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	selector         = "name=nginx-ingress-controller"
	ingressNamespace = "kube-system"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard FLAG",
	Short: "Opens the KubeApps Dashboard",
	Long:  "Opens the KubeApps Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := buildOutOfClusterConfig()
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		pods, err := clientset.CoreV1().Pods(ingressNamespace).List(metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}

		if len(pods.Items) == 0 {
			return errors.New("nginx ingress controller pod not found, run kubeapps up first")
		}

		podName := pods.Items[0].Name

		localPort, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		return runPortforward(podName, localPort)
	},
}

func runPortforward(podName string, localPort int) error {
	cmd, err := exec.LookPath("kubectl")
	if err != nil {
		return err
	}
	args := []string{"kubectl", "--namespace", ingressNamespace, "port-forward", podName, fmt.Sprintf("%d:80", localPort)}

	env := os.Environ()

	openInBrowser(fmt.Sprintf("http://localhost:%d", localPort))
	return syscall.Exec(cmd, args, env)
}

func openInBrowser(url string) error {
	args := []string{"xdg-open"}
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start()
}

func init() {
	RootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().Int("port", 8002, "local port")
}
