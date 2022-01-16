/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"fmt"

	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	informers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/signals"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd" // Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
)

type Config struct {
	MasterURL                string
	Kubeconfig               string
	RepoSyncImage            string
	RepoSyncImagePullSecrets []string
	ImagePullSecretsRefs     []corev1.LocalObjectReference
	RepoSyncCommand          string
	KubeappsNamespace        string
	GlobalReposNamespace     string
	DBURL                    string
	DBUser                   string
	DBName                   string
	DBSecretName             string
	DBSecretKey              string
	UserAgentComment         string
	Crontab                  string
	TTLSecondsAfterFinished  string
	ReposPerNamespace        bool
	CustomAnnotations        []string
	CustomLabels             []string
	ParsedCustomAnnotations  map[string]string
	ParsedCustomLabels       map[string]string
}

func Serve(serveOpts Config) error {
	cfg, err := clientcmd.BuildConfigFromFlags(serveOpts.MasterURL, serveOpts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error building kubernetes clientset: %s", err.Error())
	}

	apprepoClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error building apprepo clientset: %s", err.Error())
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// We're interested in being informed about cronjobs in kubeapps namespace only, currently.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, 0, kubeinformers.WithNamespace(serveOpts.KubeappsNamespace))
	// Enable app repo scanning to be manually set to scan the kubeapps repo only. See #1923.
	var apprepoInformerFactory informers.SharedInformerFactory
	if serveOpts.ReposPerNamespace {
		apprepoInformerFactory = informers.NewSharedInformerFactory(apprepoClient, 0)
	} else {
		apprepoInformerFactory = informers.NewFilteredSharedInformerFactory(apprepoClient, 0, serveOpts.KubeappsNamespace, nil)
	}

	controller := NewController(kubeClient, apprepoClient, kubeInformerFactory, apprepoInformerFactory, &serveOpts)

	go kubeInformerFactory.Start(stopCh)
	go apprepoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		return fmt.Errorf("Error running controller: %s", err.Error())
	}
	return nil
}
