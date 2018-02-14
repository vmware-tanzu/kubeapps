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

package kubeapps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/ksonnet/kubecfg/utils"
	"github.com/kubeapps/kubeapps/pkg/gke"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

const (
	GcTag          = "bitnami/kubeapps"
	KubeappsNS     = "kubeapps"
	KubelessNS     = "kubeless"
	SystemNS       = "kube-system"
	Kubeapps_NS    = "kubeapps"
	MongoDB_Secret = "mongodb"
)

var MongoDB_SecretFields = []string{"mongodb-root-password"}

var upCmd = &cobra.Command{
	Use:   "up FLAG",
	Short: "Install Kubeapps components.",
	Long: `Install Kubeapps components.

List of components that kubeapps up installs:

- Kubeless (https://github.com/kubeless/kubeless)
- Sealed-Secrets (https://github.com/bitnami/sealed-secrets)
- Helm/Tiller (https://github.com/kubernetes/helm)
- Kubeapps Dashboard (https://github.com/kubeapps/dashboard)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := kubecfg.UpdateCmd{
			DefaultNamespace: "default",
		}
		var err error
		c.Create = true

		c.DryRun, err = cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("can't get --dry-run flag: %v", err)
		}

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			return fmt.Errorf("can't get --out flag: %v", err)
		}

		c.GcTag = GcTag

		c.ClientPool, c.Discovery, err = restClientPool()
		if err != nil {
			return fmt.Errorf("can't get Kubernetes client: %v", err)
		}

		// validate k8s version
		version, err := utils.FetchVersion(c.Discovery)
		if err != nil {
			return fmt.Errorf("can't verify Kubernetes version: %v", err)
		}
		if version.Major <= 1 && version.Minor < 7 {
			return fmt.Errorf("kubernetes with RBAC enabled (v1.7+) is required to run Kubeapps")
		}

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return fmt.Errorf("can't read kubeapps manifest: %v", err)
		}

		objs, err := parseObjects(manifest)
		if err != nil {
			return fmt.Errorf("can't parse kubeapps manifest: %v", err)
		}

		// k8s on GKE
		if ok, err := isGKE(c.Discovery); err != nil {
			return err
		} else if ok && !c.DryRun {
			gcloudPath, err := gke.SdkConfigPath()
			if err != nil {
				return fmt.Errorf("can't get sdk config path: %v", err)
			}

			user, err := gke.GetActiveUser(gcloudPath)
			if err != nil {
				return fmt.Errorf("can't get active gke user: %v", err)
			}

			crb, err := gke.BuildCrbObject(user)
			if err != nil {
				return fmt.Errorf("can't assign cluster-admin permission to the current user: %v", err)
			}

			//(tuna): we force the deployment ordering here:
			// this clusterrolebinding will be created before others for granting the proper permission.
			// when the installation finishes, it will be gc'd immediately.
			c.SkipGc = true
			err = c.Run(crb)
			if err != nil {
				return fmt.Errorf("can't assign cluster-admin permission to the current user: %v", err)
			}
			c.SkipGc = false
		}

		// mongodb secret
		// FIXME (tuna): if the mongodb secret exists then do nothing,
		// otherwise, add it (with new generated rand pw) to the objs list then do full update.
		if prevsecret, exist, err := mongoSecretExists(c, MongoDB_Secret, Kubeapps_NS); err != nil {
			return err
		} else if !exist {
			pw := make(map[string]string)
			for _, p := range MongoDB_SecretFields {
				s, err := generateEncodedRandomPassword(12)
				if err != nil {
					return fmt.Errorf("error reading random data for secret %s: %v", MongoDB_Secret, err)
				}
				pw[p] = s
			}
			secret := buildSecretObject(pw, MongoDB_Secret, Kubeapps_NS)
			objs = append(objs, secret)
		} else if exist {
			// add prevsecret to the list so it won't be GC-ed
			objs = append(objs, prevsecret)
		}

		if c.DryRun {
			return dump(cmd.OutOrStdout(), out, c.Discovery, objs)
		}

		err = c.Run(objs)
		if err != nil {
			return fmt.Errorf("can't install kubeapps components: %v", err)
		}

		config, err := buildOutOfClusterConfig()
		if err != nil {
			return err
		}
		clientset, err := kubernetes.NewForConfig(config)

		err = printOutput(cmd.OutOrStdout(), clientset)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("dry-run", false, "Show manifest to be submitted to the k8s cluster without deploying.")
	upCmd.Flags().StringP("out", "o", "yaml", "Specify manifest format: yaml | json. Note: used only with --dry-run")
}

func dump(w io.Writer, out string, disco discovery.DiscoveryInterface, objs []*unstructured.Unstructured) error {
	bObjs := [][]byte{}
	toSort, err := utils.DependencyOrder(disco, objs)
	if err != nil {
		return fmt.Errorf("can't dump kubeapps manifest: %v", err)
	}
	sort.Sort(toSort)

	switch out {
	case "json":
		for _, obj := range objs {
			j, err := json.MarshalIndent(obj, "", "    ")
			if err != nil {
				return fmt.Errorf("can't dump kubeapps manifest: %v", err)
			}
			bObjs = append(bObjs, j)
		}

		b := bytes.Join(bObjs, []byte(fmt.Sprintf("\n")))
		fmt.Fprintln(w, string(b[:]))

	case "yaml":
		for _, obj := range objs {
			j, err := obj.MarshalJSON()
			if err != nil {
				return fmt.Errorf("can't dump kubeapps manifest: %v", err)
			}
			y, err := yaml.JSONToYAML(j)
			if err != nil {
				return err
			}
			bObjs = append(bObjs, y)
		}

		b := bytes.Join(bObjs, []byte(fmt.Sprintf("---\n")))
		fmt.Fprintln(w, string(b[:]))
	}

	return nil
}

func isGKE(disco discovery.DiscoveryInterface) (bool, error) {
	sv, err := disco.ServerVersion()
	if err != nil {
		return false, err
	}
	if strings.Contains(sv.GitVersion, "gke") {
		return true, nil
	}

	return false, nil
}

func printOutput(w io.Writer, c *kubernetes.Clientset) error {
	fmt.Printf("\nKubeapps has been deployed successfully. \n" +
		"It may take a few minutes for all components to be ready. \n\n")
	nss := []string{KubeappsNS, KubelessNS, SystemNS}
	err := printSvc(w, c, nss)
	if err != nil {
		return err
	}
	err = printDeployment(w, c, nss)
	if err != nil {
		return err
	}
	err = printStS(w, c, nss)
	if err != nil {
		return err
	}
	err = printPod(w, c, nss)
	if err != nil {
		return err
	}

	fmt.Printf("You can run `kubectl get all --all-namespaces -l created-by=kubeapps` to check the status of the Kubeapps components. \n\n")

	return nil
}

func mongoSecretExists(c kubecfg.UpdateCmd, name, ns string) (*unstructured.Unstructured, bool, error) {
	gvk := schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
	rc, err := clientForGroupVersionKind(c.ClientPool, c.Discovery, gvk, ns)
	if err != nil {
		return nil, false, err
	}
	prevSec, err := rc.Get(name, metav1.GetOptions{})

	if k8sErrors.IsNotFound(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, true, err
	}

	if prevSec.Object["data"] == nil {
		return nil, true, fmt.Errorf("secret %s already exists but it doesn't contain any expected key", name)
	}

	prevPw := prevSec.Object["data"].(map[string]interface{})
	for _, p := range MongoDB_SecretFields {
		if prevPw[p] == nil {
			return nil, true, fmt.Errorf("secret %s already exists but it doesn't contain the expected key %s", name, p)
		}
	}

	return prevSec, true, nil
}

func buildSecretObject(pw map[string]string, name, ns string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": ns,
			},
			"data": pw,
		},
	}
}

func printPod(w io.Writer, c kubernetes.Interface, nss []string) error {
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAMESPACE", "NAME", "STATUS")
	pods := []v1.Pod{}
	for _, ns := range nss {
		p, err := c.CoreV1().Pods(ns).List(metav1.ListOptions{
			LabelSelector: "created-by=kubeapps",
		})
		if err != nil {
			return err
		}
		pods = append(pods, p.Items...)
	}
	for _, p := range pods {
		table.AddRow(p.Namespace, fmt.Sprintf("pod/%s", p.Name), p.Status.Phase)
	}
	fmt.Fprintln(w, table)
	fmt.Fprintln(w)
	return nil
}

func printStS(w io.Writer, c kubernetes.Interface, nss []string) error {
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAMESPACE", "NAME", "DESIRED", "CURRENT")
	sts := []v1beta1.StatefulSet{}
	for _, ns := range nss {
		s, err := c.AppsV1beta1().StatefulSets(ns).List(metav1.ListOptions{
			LabelSelector: "created-by=kubeapps",
		})
		if err != nil {
			return err
		}
		sts = append(sts, s.Items...)
	}
	for _, s := range sts {
		table.AddRow(s.Namespace, fmt.Sprintf("statefulsets/%s", s.Name), *s.Spec.Replicas, s.Status.Replicas)
	}
	fmt.Fprintln(w, table)
	fmt.Fprintln(w)
	return nil
}

func printDeployment(w io.Writer, c kubernetes.Interface, nss []string) error {
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAMESPACE", "NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE")
	deps := []v1beta1.Deployment{}
	for _, ns := range nss {
		dep, err := c.AppsV1beta1().Deployments(ns).List(metav1.ListOptions{
			LabelSelector: "created-by=kubeapps",
		})
		if err != nil {
			return err
		}
		deps = append(deps, dep.Items...)
	}

	for _, d := range deps {
		table.AddRow(d.Namespace, fmt.Sprintf("deploy/%s", d.Name), *d.Spec.Replicas, d.Status.Replicas, d.Status.UpdatedReplicas, d.Status.AvailableReplicas)
	}
	fmt.Fprintln(w, table)
	fmt.Fprintln(w)
	return nil
}

func printSvc(w io.Writer, c kubernetes.Interface, nss []string) error {
	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = true
	table.AddRow("NAMESPACE", "NAME", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)")
	svcs := []v1.Service{}
	for _, ns := range nss {
		svc, err := c.CoreV1().Services(ns).List(metav1.ListOptions{
			LabelSelector: "created-by=kubeapps",
		})
		if err != nil {
			return err
		}
		svcs = append(svcs, svc.Items...)
	}

	for _, s := range svcs {
		eIPs := ""
		if len(s.Spec.ExternalIPs) != 0 {
			for _, ip := range s.Spec.ExternalIPs {
				eIPs = eIPs + ip + ", "
			}
		}
		ports := ""
		if len(s.Spec.Ports) != 0 {
			for _, p := range s.Spec.Ports {
				ports = ports + fmt.Sprintf("%s/%s, ", strconv.FormatInt(int64(p.Port), 10), p.Protocol)
			}
		}
		table.AddRow(s.Namespace, fmt.Sprintf("svc/%s", s.Name), s.Spec.ClusterIP, eIPs, ports)
	}
	fmt.Fprintln(w, table)
	fmt.Fprintln(w)
	return nil
}
