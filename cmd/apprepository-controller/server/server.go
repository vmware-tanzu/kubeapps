// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	clientset "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	informers "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/signals"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	APIServerURL             string
	Kubeconfig               string
	RepoSyncImage            string
	RepoSyncImagePullSecrets []string
	ImagePullSecretsRefs     []corev1.LocalObjectReference
	RepoSyncCommand          string
	KubeappsNamespace        string
	GlobalPackagingNamespace string
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
	cfg, err := clientcmd.BuildConfigFromFlags(serveOpts.APIServerURL, serveOpts.Kubeconfig)
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
