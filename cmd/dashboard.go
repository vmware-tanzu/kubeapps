/*
Copyright (c) 2017 Bitnami

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

package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	selector = "name=nginx-ingress-controller"
	// Known namespace for Kubeapps' Ingress
	ingressNamespace = "kubeapps"
	// Known port for Kubeapps' Ingress HTTP server
	ingressPort = 80
)

type dashboardCmdOptions struct {
	config    *rest.Config
	client    rest.Interface
	podName   string
	localPort int
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard FLAG",
	Short: "Opens the Kubeapps Dashboard",
	Long:  "Opens the Kubeapps Dashboard",
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

		podName := pods.Items[0].GetName()

		localPort, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		if localPort == 0 {
			localPort, err = getAvailablePort()
			if err != nil {
				return err
			}
		}

		opts := dashboardCmdOptions{config: config, client: clientset.CoreV1().RESTClient(), podName: podName, localPort: localPort}
		return opts.runPortforward()
	},
}

func (d *dashboardCmdOptions) runPortforward() error {
	stopChannel := make(chan struct{}, 1)
	readyChannel := make(chan struct{})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		if stopChannel != nil {
			close(stopChannel)
		}
	}()

	// Open the Dashboard in a browser when the port-forward is established
	go func() {
		<-readyChannel
		openInBrowser(fmt.Sprintf("http://localhost:%d", d.localPort))
	}()

	fw, err := d.newPortforwarder(stopChannel, readyChannel)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func (d *dashboardCmdOptions) newPortforwarder(stopChannel, readyChannel chan struct{}) (*portforward.PortForwarder, error) {
	req := d.client.Post().
		Resource("pods").
		Namespace(ingressNamespace).
		Name(d.podName).
		SubResource("portforward")
	url := req.URL()

	dialer, err := remotecommand.NewExecutor(d.config, "POST", url)
	if err != nil {
		return nil, err
	}

	ports := []string{fmt.Sprintf("%d:%d", d.localPort, ingressPort)}

	return portforward.New(dialer, ports, stopChannel, readyChannel, os.Stdout, os.Stderr)
}

func openInBrowser(url string) error {
	fmt.Printf("Opening %s in your default browser...\n", url)
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
	dashboardCmd.Flags().Int("port", 0, "Specify a local port for the dashboard connection.")
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}
