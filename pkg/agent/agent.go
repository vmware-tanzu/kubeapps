package agent

import (
	"errors"
	"fmt"
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

func ListReleases(actionConfig *action.Configuration, namespace string, listLimit int, status string) ([]proxy.AppOverview, error) {
	allNamespaces := namespace == ""
	cmd := action.NewList(actionConfig)
	if allNamespaces {
		cmd.AllNamespaces = true
	}
	cmd.Limit = listLimit
	releases, err := cmd.Run()
	if err != nil {
		return nil, err
	}
	appOverviews := make([]proxy.AppOverview, 0)
	for _, r := range releases {
		if allNamespaces || r.Namespace == namespace {
			appOverviews = append(appOverviews, appOverviewFromRelease(r))
		}
	}
	return appOverviews, nil
}

func NewActionConfig(driver DriverType, token, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	clientset, err := kubernetes.NewForConfig(config)
	store, err := createStorage(driver, namespace, clientset)
	if err != nil {
		return nil, err
	}
	actionConfig.RESTClientGetter = nil     // TODO replace nil with meaningful value
	actionConfig.KubeClient = kube.New(nil) // TODO replace nil with meaningful value
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof
	return actionConfig, nil
}

func createStorage(driverType DriverType, namespace string, clientset *kubernetes.Clientset) (*storage.Storage, error) {
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
		return nil, fmt.Errorf("Invalid Helm drive type: %q", driverType)
	}
	return store, nil
}

func ParseDriverType(raw string) (DriverType, error) {
	switch raw {
	case "secret", "secrets":
		return Secret, nil
	case "configmap", "configmaps":
		return ConfigMap, nil
	case "memory":
		return Memory, nil
	default:
		return Memory, errors.New("Invalid Helm driver type: " + raw)
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
