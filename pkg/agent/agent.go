// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmchartutil "helm.sh/helm/v3/pkg/chartutil"
	helmkube "helm.sh/helm/v3/pkg/kube"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8sgenericclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	k8syaml "sigs.k8s.io/yaml"
)

type AppOverview struct {
	ReleaseName   string             `json:"releaseName"`
	Version       string             `json:"version"`
	Namespace     string             `json:"namespace"`
	Icon          string             `json:"icon,omitempty"`
	Status        string             `json:"status"`
	Chart         string             `json:"chart"`
	ChartMetadata helmchart.Metadata `json:"chartMetadata"`
}

// StorageForDriver is a function type which returns a specific helmstorage.
type StorageForDriver func(namespace string, clientset *k8stypedclient.Clientset) *helmstorage.Storage

// StorageForSecrets returns a storage using the Secret helmstoragedriver.
func StorageForSecrets(namespace string, clientset *k8stypedclient.Clientset) *helmstorage.Storage {
	d := helmstoragedriver.NewSecrets(clientset.CoreV1().Secrets(namespace))
	d.Log = log.Infof
	return helmstorage.Init(d)
}

// StorageForConfigMaps returns a storage using the ConfigMap helmstoragedriver.
func StorageForConfigMaps(namespace string, clientset *k8stypedclient.Clientset) *helmstorage.Storage {
	d := helmstoragedriver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
	d.Log = log.Infof
	return helmstorage.Init(d)
}

// StorageForMemory returns a storage using the Memory helmstoragedriver.
func StorageForMemory(_ string, _ *k8stypedclient.Clientset) *helmstorage.Storage {
	d := helmstoragedriver.NewMemory()
	return helmstorage.Init(d)
}

// ListReleases lists releases in the specified namespace, or all namespaces if the empty string is given.
func ListReleases(actionConfig *helmaction.Configuration, namespace string, listLimit int, status string) ([]AppOverview, error) {
	allNamespaces := namespace == ""
	cmd := helmaction.NewList(actionConfig)
	if allNamespaces {
		cmd.AllNamespaces = true
	}
	cmd.Limit = listLimit
	if status == "all" {
		cmd.StateMask = helmaction.ListAll
	}
	releases, err := cmd.Run()
	if err != nil {
		return nil, err
	}
	appOverviews := make([]AppOverview, 0)
	for _, r := range releases {
		if allNamespaces || r.Namespace == namespace {
			appOverviews = append(appOverviews, appOverviewFromRelease(r))
		}
	}
	return appOverviews, nil
}

// CreateRelease creates a helmrelease.
func CreateRelease(actionConfig *helmaction.Configuration, name, namespace, valueString string,
	ch *helmchart.Chart, registrySecrets map[string]string, timeoutSeconds int32) (*helmrelease.Release, error) {
	// Check if the release already exists
	_, err := GetRelease(actionConfig, name)
	if err == nil {
		return nil, fmt.Errorf("release %s already exists", name)
	}
	cmd, err := newInstallCommand(actionConfig, name, namespace, registrySecrets, timeoutSeconds)
	if err != nil {
		return nil, err
	}
	values, err := getValues([]byte(valueString))
	if err != nil {
		return nil, err
	}
	release, err := cmd.Run(ch, values)
	if err != nil {
		// Simulate the Atomic flag and delete the release if failed
		errDelete := DeleteRelease(actionConfig, name, false, timeoutSeconds)
		if errDelete != nil && !strings.Contains(errDelete.Error(), "release: not found") {
			return nil, fmt.Errorf("Release %q failed: %v. Unable to delete failed release: %v", name, err, errDelete)
		}
		return nil, fmt.Errorf("Release %q failed and has been uninstalled: %v", name, err)
	}
	return release, nil
}

func newInstallCommand(actionConfig *helmaction.Configuration, name string, namespace string,
	registrySecrets map[string]string, timeoutSeconds int32) (*helmaction.Install, error) {
	cmd := helmaction.NewInstall(actionConfig)
	cmd.ReleaseName = name
	cmd.Namespace = namespace
	if timeoutSeconds > 0 {
		// Given that `cmd.Wait` is not used, this timeout will only affect pre/post hooks
		cmd.Timeout = time.Duration(timeoutSeconds) * time.Second
	}
	var err error
	cmd.PostRenderer, err = NewDockerSecretsPostRenderer(registrySecrets)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// UpgradeRelease upgrades a helmrelease.
func UpgradeRelease(actionConfig *helmaction.Configuration, name, valuesYaml string,
	ch *helmchart.Chart, registrySecrets map[string]string, timeoutSeconds int32) (*helmrelease.Release, error) {
	// Check if the release already exists:
	_, err := GetRelease(actionConfig, name)
	if err != nil {
		return nil, err
	}
	log.Printf("Upgrading release %s", name)
	cmd := helmaction.NewUpgrade(actionConfig)
	if timeoutSeconds > 0 {
		// Given that `cmd.Wait` is not used, this timeout will only affect pre/post hooks
		cmd.Timeout = time.Duration(timeoutSeconds) * time.Second
	}

	cmd.PostRenderer, err = NewDockerSecretsPostRenderer(registrySecrets)
	if err != nil {
		return nil, err
	}
	values, err := helmchartutil.ReadValues([]byte(valuesYaml))
	if err != nil {
		return nil, fmt.Errorf("Unable to upgrade the release because values could not be parsed: %v", err)
	}
	res, err := cmd.Run(name, ch, values)
	if err != nil {
		return nil, fmt.Errorf("Unable to upgrade the release: %v", err)
	}
	return res, nil
}

// RollbackRelease rolls back a release to the specified revision.
func RollbackRelease(actionConfig *helmaction.Configuration, releaseName string, revision int, timeoutSeconds int32) (*helmrelease.Release, error) {
	log.Printf("Rolling back %s to revision %d.", releaseName, revision)
	rollback := helmaction.NewRollback(actionConfig)
	rollback.Version = revision
	if timeoutSeconds > 0 {
		// Given that `rollback.Wait` is not used, this timeout will only affect pre/post hooks
		rollback.Timeout = time.Duration(timeoutSeconds) * time.Second
	}
	err := rollback.Run(releaseName)
	if err != nil {
		return nil, err
	}

	// The Helm 3 rollback action does not return the new release, unlike the Helm 2 equivalent,
	// so we grab it explicitly as it's required by Kubeapps.
	return GetRelease(actionConfig, releaseName)
}

// GetRelease returns the info of a helmrelease.
func GetRelease(actionConfig *helmaction.Configuration, name string) (*helmrelease.Release, error) {
	// Namespace is already known by the RESTClientGetter.
	cmd := helmaction.NewGet(actionConfig)
	release, err := cmd.Run(name)
	if err != nil {
		return nil, err
	}
	return release, nil
}

// DeleteRelease deletes a helmrelease.
func DeleteRelease(actionConfig *helmaction.Configuration, name string, keepHistory bool, timeoutSeconds int32) error {
	// Namespace is already known by the RESTClientGetter.
	cmd := helmaction.NewUninstall(actionConfig)
	cmd.KeepHistory = keepHistory
	if timeoutSeconds > 0 {
		// Given that `cmd.Wait` is not used, this timeout will only affect pre/post hooks
		cmd.Timeout = time.Duration(timeoutSeconds) * time.Second
	}
	_, err := cmd.Run(name)
	return err
}

// NewActionConfig creates an helmaction.Configuration, which can then be used to create Helm 3 actions.
// Among other things, the helmaction.Configuration controls which namespace the command is run against.
func NewActionConfig(storageForDriver StorageForDriver, config *k8srest.Config, clientset *k8stypedclient.Clientset, namespace string) (*helmaction.Configuration, error) {
	actionConfig := new(helmaction.Configuration)
	store := storageForDriver(namespace, clientset)
	restClientGetter := NewConfigFlagsFromCluster(namespace, config)
	actionConfig.RESTClientGetter = restClientGetter
	actionConfig.KubeClient = helmkube.New(restClientGetter)
	actionConfig.Releases = store
	actionConfig.Log = log.Infof
	return actionConfig, nil
}

// NewConfigFlagsFromCluster returns ConfigFlags with default values set from within cluster.
func NewConfigFlagsFromCluster(namespace string, clusterConfig *k8srest.Config) k8sgenericclioptions.RESTClientGetter {
	impersonateGroup := []string{}

	// CertFile and KeyFile must be nil for the BearerToken to be used for authentication and authorization instead of the pod's service account.
	configFlags := &k8sgenericclioptions.ConfigFlags{
		Insecure:         &clusterConfig.TLSClientConfig.Insecure,
		Timeout:          stringptr("0"),
		Namespace:        stringptr(namespace),
		APIServer:        stringptr(clusterConfig.Host),
		CAFile:           stringptr(clusterConfig.CAFile),
		BearerToken:      stringptr(clusterConfig.BearerToken),
		ImpersonateGroup: &impersonateGroup,
	}
	return &configForCluster{
		config:         clusterConfig,
		discoveryBurst: 100,
		ConfigFlags:    configFlags,
	}
}

// Values is a type alias for values.yaml.
type Values map[string]interface{}

func getValues(raw []byte) (Values, error) {
	values := make(Values)
	err := k8syaml.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func stringptr(val string) *string {
	return &val
}

// ParseDriverType maps strings to well-typed driver representations.
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

func appOverviewFromRelease(r *helmrelease.Release) AppOverview {
	return AppOverview{
		ReleaseName:   r.Name,
		Version:       r.Chart.Metadata.Version,
		Icon:          r.Chart.Metadata.Icon,
		Namespace:     r.Namespace,
		Status:        r.Info.Status.String(),
		Chart:         r.Chart.Name(),
		ChartMetadata: *r.Chart.Metadata,
	}
}
