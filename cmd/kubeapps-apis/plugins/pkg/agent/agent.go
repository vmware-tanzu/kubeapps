// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"fmt"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
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

// CreateRelease creates a release.
func CreateRelease(actionConfig *action.Configuration, name, namespace, valueString string,
	ch *chart.Chart, registrySecrets map[string]string, timeoutSeconds int32) (*release.Release, error) {
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

func newInstallCommand(actionConfig *action.Configuration, name string, namespace string,
	registrySecrets map[string]string, timeoutSeconds int32) (*action.Install, error) {
	cmd := action.NewInstall(actionConfig)
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

// UpgradeRelease upgrades a release.
func UpgradeRelease(actionConfig *action.Configuration, name, valuesYaml string,
	ch *chart.Chart, registrySecrets map[string]string, timeoutSeconds int32) (*release.Release, error) {
	// Check if the release already exists:
	_, err := GetRelease(actionConfig, name)
	if err != nil {
		return nil, err
	}
	log.InfoS("Upgrading release", "release-name", name)
	cmd := action.NewUpgrade(actionConfig)
	if timeoutSeconds > 0 {
		// Given that `cmd.Wait` is not used, this timeout will only affect pre/post hooks
		cmd.Timeout = time.Duration(timeoutSeconds) * time.Second
	}

	cmd.PostRenderer, err = NewDockerSecretsPostRenderer(registrySecrets)
	if err != nil {
		return nil, err
	}
	values, err := chartutil.ReadValues([]byte(valuesYaml))
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
func RollbackRelease(actionConfig *action.Configuration, releaseName string, revision int, timeoutSeconds int32) (*release.Release, error) {
	log.Infof("Rolling back %s to revision %d.", releaseName, revision)
	rollback := action.NewRollback(actionConfig)
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

// GetRelease returns the info of a release.
func GetRelease(actionConfig *action.Configuration, name string) (*release.Release, error) {
	// Namespace is already known by the RESTClientGetter.
	cmd := action.NewGet(actionConfig)
	release, err := cmd.Run(name)
	if err != nil {
		return nil, err
	}
	return release, nil
}

// DeleteRelease deletes a release.
func DeleteRelease(actionConfig *action.Configuration, name string, keepHistory bool, timeoutSeconds int32) error {
	// Namespace is already known by the RESTClientGetter.
	cmd := action.NewUninstall(actionConfig)
	cmd.KeepHistory = keepHistory
	if timeoutSeconds > 0 {
		// Given that `cmd.Wait` is not used, this timeout will only affect pre/post hooks
		cmd.Timeout = time.Duration(timeoutSeconds) * time.Second
	}
	_, err := cmd.Run(name)
	return err
}

// NewActionConfig creates an action.Configuration, which can then be used to create Helm 3 actions.
// Among other things, the action.Configuration controls which namespace the command is run against.
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

// NewConfigFlagsFromCluster returns ConfigFlags with default values set from within cluster.
func NewConfigFlagsFromCluster(namespace string, clusterConfig *rest.Config) genericclioptions.RESTClientGetter {
	impersonateGroup := []string{}

	// CertFile and KeyFile must be nil for the BearerToken to be used for authentication and authorization instead of the pod's service account.
	configFlags := &genericclioptions.ConfigFlags{
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
	err := yaml.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func stringptr(val string) *string {
	return &val
}
