// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	apprepoclient "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	apprepoinformers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	appreposignals "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/signals"
	k8scorev1 "k8s.io/api/core/v1"
	k8sinformers "k8s.io/client-go/informers"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stoolsclientcmd "k8s.io/client-go/tools/clientcmd" // Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
)

type Config struct {
	APIServerURL             string
	Kubeconfig               string
	RepoSyncImage            string
	RepoSyncImagePullSecrets []string
	ImagePullSecretsRefs     []k8scorev1.LocalObjectReference
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
	cfg, err := k8stoolsclientcmd.BuildConfigFromFlags(serveOpts.APIServerURL, serveOpts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := k8stypedclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error building kubernetes clientset: %s", err.Error())
	}

	apprepoClient, err := apprepoclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error building apprepo clientset: %s", err.Error())
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := appreposignals.SetupSignalHandler()

	// We're interested in being informed about cronjobs in kubeapps namespace only, currently.
	kubeInformerFactory := k8sinformers.NewSharedInformerFactoryWithOptions(kubeClient, 0, k8sinformers.WithNamespace(serveOpts.KubeappsNamespace))
	// Enable app repo scanning to be manually set to scan the kubeapps repo only. See #1923.
	var apprepoInformerFactory apprepoinformers.SharedInformerFactory
	if serveOpts.ReposPerNamespace {
		apprepoInformerFactory = apprepoinformers.NewSharedInformerFactory(apprepoClient, 0)
	} else {
		apprepoInformerFactory = apprepoinformers.NewFilteredSharedInformerFactory(apprepoClient, 0, serveOpts.KubeappsNamespace, nil)
	}

	controller := NewController(kubeClient, apprepoClient, kubeInformerFactory, apprepoInformerFactory, &serveOpts)

	go kubeInformerFactory.Start(stopCh)
	go apprepoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		return fmt.Errorf("Error running controller: %s", err.Error())
	}
	return nil
}
