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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/ksonnet/kubecfg/utils"
	"github.com/spf13/cobra"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeapps/kubeapps/pkg/gke"
	yamlUtils "github.com/kubeapps/kubeapps/pkg/yaml"
)

const (
	GcTag          = "bitnami/kubeapps"
	KubeappsNS     = "kubeapps"
	KubelessNS     = "kubeless"
	SystemNS       = "kube-system"
	Kubeapps_NS    = "kubeapps"
	MongoDB_Secret = "mongodb"
	tillerSecret   = "tiller-secret"
)

var MongoDB_SecretFields = []string{"mongodb-root-password"}

var (
	tlsCaCertDefault = fmt.Sprintf("%s/ca.crt", os.Getenv("HELM_HOME"))
	tlsCertDefault   = fmt.Sprintf("%s/tls.crt", os.Getenv("HELM_HOME"))
	tlsKeyDefault    = fmt.Sprintf("%s/tls.key", os.Getenv("HELM_HOME"))

	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsVerify     bool   // enable TLS and verify remote certificates
	tlsEnable     bool   // enable TLS
)

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
		if version.Major <= 1 && version.Minor < 8 {
			return fmt.Errorf("kubernetes v1.8+ is required to run Kubeapps")
		}

		manifest, err := fsGetFile("/kubeapps-objs.yaml")
		if err != nil {
			return fmt.Errorf("can't read kubeapps manifest: %v", err)
		}

		objs, err := yamlUtils.ParseObjects(manifest)
		if err != nil {
			return fmt.Errorf("can't parse kubeapps manifest: %v", err)
		}

		// k8s on GKE
		if ok, err := isGKE(c.Discovery); err != nil {
			return err
		} else if ok && !c.DryRun {
			user, err := gke.GetActiveUser()
			if err != nil {
				return err
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

		// TLS Configuration
		if tlsVerify || tlsEnable {
			// Generate secret
			tlsCaCert, err := ioutil.ReadFile(tlsCaCertFile)
			if err != nil {
				return err
			}
			tlsCert, err := ioutil.ReadFile(tlsCertFile)
			if err != nil {
				return err
			}
			tlsKey, err := ioutil.ReadFile(tlsKeyFile)
			if err != nil {
				return err
			}
			data := map[string]string{
				"ca.crt":  base64.StdEncoding.EncodeToString(tlsCaCert),
				"tls.crt": base64.StdEncoding.EncodeToString(tlsCert),
				"tls.key": base64.StdEncoding.EncodeToString(tlsKey),
			}
			tlsCertSecret := buildSecretObject(data, tillerSecret, Kubeapps_NS)
			objs = append(objs, tlsCertSecret)
			// Modify tiller deployment to use TLS
			unstructuredTillerDeployment, index, err := findObj(objs, "tiller-deploy", Kubeapps_NS, "Deployment")
			if err != nil {
				return err
			}
			depBytes, err := json.Marshal(unstructuredTillerDeployment.Object)
			if err != nil {
				return err
			}
			tillerDeployment := v1beta1.Deployment{}
			err = json.Unmarshal(depBytes, &tillerDeployment)
			if err != nil {
				return err
			}
			certPath := "/etc/certs"
			modifiedContainers := []v1.Container{}
			for _, c := range tillerDeployment.Spec.Template.Spec.Containers {
				switch c.Name {
				case "tiller":
					c.Env = append(
						c.Env,
						v1.EnvVar{Name: "TILLER_TLS_ENABLE", Value: "1"},
						v1.EnvVar{Name: "TILLER_TLS_CERTS", Value: certPath},
					)
					if tlsVerify {
						c.Env = append(c.Env, v1.EnvVar{Name: "TILLER_TLS_VERIFY", Value: "1"})
					}
				case "proxy":
					c.Args = append(c.Args, "--tls")
					if tlsVerify {
						c.Args = append(c.Args, "--tls-verify")
					}
					c.Env = append(c.Env, v1.EnvVar{Name: "HELM_HOME", Value: certPath})
				default:
					return fmt.Errorf("Unexpected container %s", c.Name)
				}
				c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
					Name:      "tiller-certs",
					MountPath: certPath,
					ReadOnly:  true,
				})
				modifiedContainers = append(modifiedContainers, c)
			}
			tillerDeployment.Spec.Template.Spec.Containers = modifiedContainers
			tillerDeployment.Spec.Template.Spec.Volumes = append(
				tillerDeployment.Spec.Template.Spec.Volumes,
				v1.Volume{
					Name: "tiller-certs",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: tillerSecret,
						},
					},
				},
			)
			newDpm, err := json.Marshal(tillerDeployment)
			if err != nil {
				return err
			}
			err = json.Unmarshal(newDpm, unstructuredTillerDeployment)
			if err != nil {
				return err
			}
			objs[index] = unstructuredTillerDeployment
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

func findObj(objs []*unstructured.Unstructured, name, namespace, kind string) (*unstructured.Unstructured, int, error) {
	index := 0
	for _, obj := range objs {
		if obj.GetName() == name && obj.GetNamespace() == namespace && obj.GetKind() == kind {
			return obj, index, nil
		}
		index++
	}
	return nil, -1, fmt.Errorf("Obj %s/%s (%s) not found", namespace, name, kind)
}

func init() {
	RootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("dry-run", false, "Show manifest to be submitted to the k8s cluster without deploying.")
	upCmd.Flags().StringP("out", "o", "yaml", "Specify manifest format: yaml | json. Note: used only with --dry-run")
	// TLS Flags
	upCmd.Flags().StringVar(&tlsCaCertFile, "tls-ca-cert", tlsCaCertDefault, "path to TLS CA certificate file")
	upCmd.Flags().StringVar(&tlsCertFile, "tiller-tls-cert", tlsCertDefault, "path to TLS certificate file")
	upCmd.Flags().StringVar(&tlsKeyFile, "tiller-tls-key", tlsKeyDefault, "path to TLS key file")
	upCmd.Flags().BoolVar(&tlsVerify, "tiller-tls-verify", false, "enable TLS for request and verify remote")
	upCmd.Flags().BoolVar(&tlsEnable, "tiller-tls", false, "enable TLS for request")
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
				"labels": map[string]string{
					"created-by": "kubeapps",
				},
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
