package agent

import (
	"errors"
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

// StorageForDriver is a function type which returns a specific storage.
type StorageForDriver func(namespace string, clientset *kubernetes.Clientset) *storage.Storage

// StorageForSecrets returns a storage using the Secret driver.
func StorageForSecrets(namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
	d.Log = klog.Infof
	return storage.Init(d)
}

// StorageForConfigMaps returns a storage using the ConfigMap driver.
func StorageForConfigMaps(namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
	d.Log = klog.Infof
	return storage.Init(d)
}

// StorageForMemory returns a storage using the Memory driver.
func StorageForMemory(_ string, _ *kubernetes.Clientset) *storage.Storage {
	d := driver.NewMemory()
	return storage.Init(d)
}

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

func NewActionConfig(storageForDriver StorageForDriver, token, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	clientset, err := kubernetes.NewForConfig(config)
	store := storageForDriver(namespace, clientset)
	actionConfig.RESTClientGetter = nil     // TODO replace nil with meaningful value
	actionConfig.KubeClient = kube.New(nil) // TODO replace nil with meaningful value
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof
	return actionConfig, nil
}

func ParseDriverType(raw string) (StorageForDriver, error) {
	switch raw {
	case "secret", "secrets":
		return StorageForSecrets, nil
	case "configmap", "configmaps":
		return StorageForConfigMaps, nil
	case "memory":
		return StorageForMemory, nil
	default:
		return nil, errors.New("Invalid Helm driver type: " + raw)
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
