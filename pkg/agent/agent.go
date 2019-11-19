package agent

import (
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

type DriverType string

const (
	Secret    DriverType = "SECRET"
	ConfigMap DriverType = "CONFIGMAP"
	Memory    DriverType = "MEMORY"
)

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

func NewActionConfig(driver DriverType, token, namespace string) *action.Configuration {
	actionConfig := new(action.Configuration)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	clientset, err := kubernetes.NewForConfig(config)
	store := createStorage(driver, namespace, clientset)
	actionConfig.RESTClientGetter = nil     // TODO replace nil with meaningful value
	actionConfig.KubeClient = kube.New(nil) // TODO replace nil with meaningful value
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof
	return actionConfig
}

func createStorage(driverType DriverType, namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	var store *storage.Storage
	switch driverType {
	case Secret:
		d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case ConfigMap:
		d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case Memory:
		d := driver.NewMemory()
		store = storage.Init(d)
	default:
		// No (real) enums/ADTs in Go, so no static guarantee against this case.
		panic("Invalid Helm driver type: " + driverType)
	}
	return store
}

func ParseDriverType(raw string) DriverType {
	switch raw {
	case "secret", "secrets":
		return Secret
	case "configmap", "configmaps":
		return ConfigMap
	case "memory":
		return Memory
	default:
		panic("Invalid Helm driver type: " + raw)
	}
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
