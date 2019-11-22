package agent

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kubeapps/kubeapps/pkg/proxy"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const driverEnvVar = "HELM_DRIVER"

type Options struct {
	ListLimit int
	Timeout   int64
}

type Config struct {
	ActionConfig *action.Configuration
	AgentOptions Options
}

func ListReleases(config Config, namespace string, status string) ([]proxy.AppOverview, error) {
	cmd := action.NewList(config.ActionConfig)
	if namespace == "" {
		cmd.AllNamespaces = true
	}
	cmd.Limit = config.AgentOptions.ListLimit
	releases, err := cmd.Run()
	if err != nil {
		return nil, err
	}
	appOverviews := make([]proxy.AppOverview, len(releases))
	for i, r := range releases {
		appOverviews[i] = appOverviewFromRelease(r)
	}
	return appOverviews, nil
}

func NewActionConfig(token, namespace string) *action.Configuration {
	actionConfig := new(action.Configuration)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	clientset, err := kubernetes.NewForConfig(config)
	store := createStorage(os.Getenv(driverEnvVar), namespace, clientset)
	actionConfig.RESTClientGetter = nil
	actionConfig.KubeClient = kube.New(nil)
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof
	return actionConfig
}

func createStorage(driverType, namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	var store *storage.Storage
	switch driverType {
	case "", "secret", "secrets":
		d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "configmap", "configmaps":
		d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "memory":
		d := driver.NewMemory()
		store = storage.Init(d)
	default:
		// Not sure what to do here.
		panic(fmt.Sprintf("Unknown value of environment variable %s: %s", driverEnvVar, driverType))
	}
	return store
}

func appOverviewFromRelease(r *release.Release) proxy.AppOverview {
	return proxy.AppOverview{
		ReleaseName: r.Name,
		Version:     strconv.Itoa(r.Version),
		Icon:        r.Chart.Metadata.Icon,
		Namespace:   r.Namespace,
		Status:      r.Info.Status.String(),
	}
}
