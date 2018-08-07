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
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	tillerSelector = "OWNER=TILLER"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate-configmaps-to-secrets",
	Short: "Migrates Helm v2 releases from ConfigMaps to Secrets",
	Long:  "Migrates Helm v2 releases from ConfigMaps to Secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		tillerNs, err := cmd.Flags().GetString("target-tiller-namespace")
		if err != nil {
			return fmt.Errorf("can't get --target-tiller-namespace flag: %v", err)
		}

		config, err := buildOutOfClusterConfig()
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		configMaps, err := clientset.CoreV1().ConfigMaps(Kubeapps_NS).List(metav1.ListOptions{LabelSelector: tillerSelector})
		if err != nil {
			return err
		}

		if len(configMaps.Items) == 0 {
			return errors.New("No releases found")
		}

		for _, cm := range configMaps.Items {
			_, err := clientset.CoreV1().Secrets(tillerNs).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm.Name,
					Namespace: tillerNs,
					Labels:    cm.Labels,
				},
				Data: map[string][]byte{
					"release": []byte(cm.Data["release"]),
				},
			})
			if err != nil {
				if k8sErrors.IsAlreadyExists(err) {
					log.Printf("Skipping release %s since it already exists as secret", cm.Name)
				} else {
					return fmt.Errorf("Unable to create secret: %v", err)
				}
			} else {
				log.Printf("Migrated %s as a secret", cm.Name)
			}
		}
		log.Printf("Done. ConfigMaps are left in the namespace %s to debug possible errors. Please delete them manually", Kubeapps_NS)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().String("target-tiller-namespace", "kube-system", "Namespace of target Tiller.")
}
