package agent

import (
	"errors"
	"strconv"

	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

// StorageForDriver is a function type which returns a specific storage.
type StorageForDriver func(namespace string, clientset *kubernetes.Clientset) *storage.Storage

// StorageForSecrets returns a storage using the Secret driver.
func StorageForSecrets(namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
	d.Log = log.Infof
	return storage.Init(d)
}

// StorageForConfigMaps returns a storage using the ConfigMap driver.
func StorageForConfigMaps(namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
	d.Log = log.Infof
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
	UserAgent string
}

type Config struct {
	ActionConfig *action.Configuration
	AgentOptions Options
	ChartClient  chartUtils.Resolver
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

func CreateRelease(config Config, name, namespace, valueString string, ch *chart.Chart) (*release.Release, error) {
	cmd := action.NewInstall(config.ActionConfig)
	cmd.ReleaseName = name
	cmd.Namespace = namespace
	values, err := getValues([]byte(valueString))
	if err != nil {
		return nil, err
	}
	release, err := cmd.Run(ch, values)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func GetRelease(actionConfig *action.Configuration, name string) (*release.Release, error) {
	// Namespace is already known by the RESTClientGetter.
	cmd := action.NewGet(actionConfig)
	release, err := cmd.Run(name)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func DeleteRelease(actionConfig *action.Configuration, name string, keepHistory bool) error {
	// Namespace is already known by the RESTClientGetter.
	cmd := action.NewUninstall(actionConfig)
	cmd.KeepHistory = keepHistory
	_, err := cmd.Run(name)
	return err
}

func NewActionConfig(storageForDriver StorageForDriver, config *rest.Config, clientset *kubernetes.Clientset, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	store := storageForDriver(namespace, clientset)
	restClientGetter := NewConfigFlagsFromCluster(namespace, config)
	actionConfig.RESTClientGetter = restClientGetter
	actionConfig.KubeClient = kube.New(restClientGetter)
	actionConfig.Releases = store
	actionConfig.Log = log.Infof
	return actionConfig, nil
}

// NewConfigFlagsFromCluster returns ConfigFlags with default values set from within cluster
func NewConfigFlagsFromCluster(namespace string, clusterConfig *rest.Config) *genericclioptions.ConfigFlags {
	impersonateGroup := []string{}
	insecure := false

	// CertFile and KeyFile must be nil for the BearerToken to be used for authentication and authorization instead of the pod's service account.
	return &genericclioptions.ConfigFlags{
		Insecure:         &insecure,
		Timeout:          stringptr("0"),
		Namespace:        stringptr(namespace),
		APIServer:        stringptr(clusterConfig.Host),
		CAFile:           stringptr(clusterConfig.CAFile),
		BearerToken:      stringptr(clusterConfig.BearerToken),
		ImpersonateGroup: &impersonateGroup,
	}
}

// Values is a type alias for values.yaml
type Values map[string]interface{}

func getValues(raw []byte) (Values, error) {
	values := make(Values)
	err := yaml.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func stringptr(val string) *string {
	return &val
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
