/*
Copyright 2017 Bitnami.

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

package main

import (
	"flag"

	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	informers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd" // Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).

	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	log "github.com/sirupsen/logrus"
)

var (
	masterURL         string
	kubeconfig        string
	repoSyncImage     string
	repoSyncCommand   string
	namespace         string
	dbType            string
	dbURL             string
	dbUser            string
	dbName            string
	dbSecretName      string
	dbSecretKey       string
	userAgentComment  string
	crontab           string
	reposPerNamespace bool
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	apprepoClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building apprepo clientset: %s", err.Error())
	}

	// We're interested in being informed about cronjobs in kubeapps namespace only, currently.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, 0, kubeinformers.WithNamespace(namespace))
	apprepoInformerFactory := informers.NewSharedInformerFactory(apprepoClient, 0)

	controller := NewController(kubeClient, apprepoClient, kubeInformerFactory, apprepoInformerFactory, namespace)

	go kubeInformerFactory.Start(stopCh)
	go apprepoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&repoSyncImage, "repo-sync-image", "quay.io/helmpack/chart-repo:latest", "container repo/image to use in CronJobs")
	flag.StringVar(&repoSyncCommand, "repo-sync-cmd", "/chart-repo", "command used to sync/delete repos for repo-sync-image")
	flag.StringVar(&namespace, "namespace", "kubeapps", "Namespace to discover AppRepository resources")
	flag.BoolVar(&reposPerNamespace, "repos-per-namespace", true, "UNUSED: This flag will be removed in a future release.")
	flag.StringVar(&dbType, "database-type", "mongodb", "Database type. Allowed values: mongodb, postgresql")
	flag.StringVar(&dbURL, "database-url", "localhost", "Database URL")
	flag.StringVar(&dbUser, "database-user", "root", "Database user")
	flag.StringVar(&dbName, "database-name", "charts", "Database name")
	flag.StringVar(&dbSecretName, "database-secret-name", "mongodb", "Kubernetes secret name for database credentials")
	flag.StringVar(&dbSecretKey, "database-secret-key", "mongodb-root-password", "Kubernetes secret key used for database credentials")
	flag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	flag.StringVar(&crontab, "crontab", "*/10 * * * *", "CronTab to specify schedule")
}
