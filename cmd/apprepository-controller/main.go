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
	"bytes"
	"os"

	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	informers "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/informers/externalversions"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/signals"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd" // Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).

	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// Config contains all the flags passed through the command line
// besides, it contains the []LocalObjectReference for the PullSecrets
// so they can be shared across the jobs
type Config struct {
	MasterURL                string
	Kubeconfig               string
	RepoSyncImage            string
	RepoSyncImagePullSecrets []string
	ImagePullSecretsRefs     []corev1.LocalObjectReference
	RepoSyncCommand          string
	KubeappsNamespace        string
	DBURL                    string
	DBUser                   string
	DBName                   string
	DBSecretName             string
	DBSecretKey              string
	UserAgentComment         string
	Crontab                  string
	TTLSecondsAfterFinished  string
	ReposPerNamespace        bool

	// Args are the positional (non-flag) command-line arguments.
	Args []string
}

// parseFlags parses the command-line arguments provided to the program.
// Typically os.Args[0] is provided as 'programName' and os.args[1:] as 'args'.
func parseFlags(progname string, args []string) (config *Config, output string, err error) {
	flagSet := flag.NewFlagSet(progname, flag.ContinueOnError)
	var buf bytes.Buffer
	flagSet.SetOutput(&buf)

	var conf Config

	flagSet.StringVar(&conf.Kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flagSet.StringVar(&conf.MasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flagSet.StringVar(&conf.RepoSyncImage, "repo-sync-image", "quay.io/helmpack/chart-repo:latest", "container repo/image to use in CronJobs")
	flagSet.StringSliceVar(&conf.RepoSyncImagePullSecrets, "repo-sync-image-pullsecrets", nil, "optional reference to secrets in the same namespace to use for pulling the image used by this pod")
	flagSet.StringVar(&conf.RepoSyncCommand, "repo-sync-cmd", "/chart-repo", "command used to sync/delete repos for repo-sync-image")
	flagSet.StringVar(&conf.KubeappsNamespace, "namespace", "kubeapps", "Namespace to discover AppRepository resources")
	flagSet.BoolVar(&conf.ReposPerNamespace, "repos-per-namespace", true, "Defaults to watch for repos in all namespaces. Switch to false to watch only the configured namespace.")
	flagSet.StringVar(&conf.DBURL, "database-url", "localhost", "Database URL")
	flagSet.StringVar(&conf.DBUser, "database-user", "root", "Database user")
	flagSet.StringVar(&conf.DBName, "database-name", "charts", "Database name")
	flagSet.StringVar(&conf.DBSecretName, "database-secret-name", "kubeapps-db", "Kubernetes secret name for database credentials")
	flagSet.StringVar(&conf.DBSecretKey, "database-secret-key", "postgresql-root-password", "Kubernetes secret key used for database credentials")
	flagSet.StringVar(&conf.UserAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	flagSet.StringVar(&conf.Crontab, "crontab", "*/10 * * * *", "CronTab to specify schedule")
	// TTLSecondsAfterFinished specifies the number of seconds a sync job should live after finishing.
	// The support for this is currently alpha in K8s itself, requiring a feature gate being set to enable
	// it. See https://kubernetes.io/docs/concepts/workloads/controllers/job/#clean-up-finished-jobs-automatically
	flagSet.StringVar(&conf.TTLSecondsAfterFinished, "ttl-lifetime-afterfinished-job", "3600", "Lifetime limit after which the resource Jobs are deleted expressed in seconds by default is 3600 (1h) ")

	err = flagSet.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}
	conf.Args = flagSet.Args()
	return &conf, buf.String(), nil
}

func main() {
	conf, output, err := parseFlags(os.Args[0], os.Args[1:])

	if err != nil {
		log.Fatal("error parsing command-line arguments: ", err)
		log.Fatal(output)
		log.Exit(1)
	}

	log.WithFields(log.Fields{
		"kubeconfig":                     conf.Kubeconfig,
		"master":                         conf.MasterURL,
		"repo-sync-image":                conf.RepoSyncImage,
		"repo-sync-image-pullsecrets":    conf.RepoSyncImagePullSecrets,
		"repo-sync-cmd":                  conf.RepoSyncCommand,
		"namespace":                      conf.KubeappsNamespace,
		"repos-per-namespace":            conf.ReposPerNamespace,
		"database-url":                   conf.DBURL,
		"database-user":                  conf.DBUser,
		"database-name":                  conf.DBName,
		"database-secret-name":           conf.DBSecretName,
		"database-secret-key":            conf.DBSecretKey,
		"user-agent-comment":             conf.UserAgentComment,
		"crontab":                        conf.Crontab,
		"ttl-lifetime-afterfinished-job": conf.TTLSecondsAfterFinished,
	}).Info("apprepository-controller configured with these args:")

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(conf.MasterURL, conf.Kubeconfig)
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
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, 0, kubeinformers.WithNamespace(conf.KubeappsNamespace))
	// Enable app repo scanning to be manually set to scan the kubeapps repo only. See #1923.
	var apprepoInformerFactory informers.SharedInformerFactory
	if conf.ReposPerNamespace {
		apprepoInformerFactory = informers.NewSharedInformerFactory(apprepoClient, 0)
	} else {
		apprepoInformerFactory = informers.NewFilteredSharedInformerFactory(apprepoClient, 0, conf.KubeappsNamespace, nil)
	}

	conf.ImagePullSecretsRefs = getImagePullSecretsRefs(conf.RepoSyncImagePullSecrets)

	controller := NewController(kubeClient, apprepoClient, kubeInformerFactory, apprepoInformerFactory, conf)

	go kubeInformerFactory.Start(stopCh)
	go apprepoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}

// getImagePullSecretsRefs gets the []string of Secrets names from the
// StringSliceVar flag list passed in the repoSyncImagePullSecrets arg
func getImagePullSecretsRefs(imagePullSecretsRefsArr []string) []corev1.LocalObjectReference {
	var imagePullSecretsRefs []corev1.LocalObjectReference

	// getting and appending a []LocalObjectReference for each ImagePullSecret passed
	for _, imagePullSecretName := range imagePullSecretsRefsArr {
		imagePullSecretsRefs = append(imagePullSecretsRefs, corev1.LocalObjectReference{Name: imagePullSecretName})
	}
	return imagePullSecretsRefs
}
