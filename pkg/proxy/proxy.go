/*
Copyright (c) 2018 Bitnami

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

package proxy

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	appMutex        map[string]*sync.Mutex
	releaseStatuses []release.Status_Code
)

func init() {
	appMutex = make(map[string]*sync.Mutex)
	releaseStatuses = []release.Status_Code{
		release.Status_UNKNOWN,
		release.Status_DEPLOYED,
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_UPGRADE,
		release.Status_PENDING_ROLLBACK,
	}
}

// Proxy contains all the elements to contact Tiller and the K8s API
type Proxy struct {
	kubeClient kubernetes.Interface
	helmClient helm.Interface
	listLimit  int
}

func isNotFound(err error) bool {
	// Ideally this would be `grpc.Code(err) == codes.NotFound`,
	// but it seems helm doesn't return grpc codes
	return strings.Contains(grpc.ErrorDesc(err), "not found")
}

// NewProxy creates a Proxy
func NewProxy(kubeClient kubernetes.Interface, helmClient helm.Interface) *Proxy {
	return &Proxy{
		kubeClient: kubeClient,
		helmClient: helmClient,
	}
}

// AppOverview represents the basics of a release
type AppOverview struct {
	ReleaseName string `json:"releaseName"`
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
	Icon        string `json:"icon,omitempty"`
	Status      string `json:"status"`
}

func (p *Proxy) get(name, namespace string) (*release.Release, error) {
	list, err := p.helmClient.ListReleases(
		helm.ReleaseListFilter(name),
		helm.ReleaseListNamespace(namespace),
		helm.ReleaseListStatuses(releaseStatuses),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to list helm releases: %v", err)
	}
	var rel *release.Release
	if list != nil && list.Releases != nil {
		for _, r := range list.Releases {
			if (namespace == "" || namespace == r.Namespace) && r.Name == name {
				rel = r
				break
			}
		}
	}
	if rel == nil {
		return nil, fmt.Errorf("Release %s not found in namespace %s", name, namespace)
	}
	return rel, nil
}

// GetReleaseStatus prints the status of the given release if exists
func (p *Proxy) GetReleaseStatus(relName string) (release.Status_Code, error) {
	status, err := p.helmClient.ReleaseStatus(relName)
	if err == nil {
		if status.Info != nil && status.Info.Status != nil {
			return status.Info.Status.Code, nil
		}
	}
	return release.Status_Code(0), fmt.Errorf("Unable to fetch release status for %s: %v", relName, err)
}

// ResolveManifest returns a manifest given the chart parameters
func (p *Proxy) ResolveManifest(namespace, values string, ch *chart.Chart) (string, error) {
	// We use the release returned after running a dry-run to know the elements to install
	resDry, err := p.helmClient.InstallReleaseFromChart(
		ch,
		namespace,
		helm.ValueOverrides([]byte(values)),
		helm.ReleaseName(""),
		helm.InstallDryRun(true),
	)
	if err != nil {
		return "", err
	}
	// The manifest returned has some extra new lines at the beginning
	return strings.TrimLeft(resDry.Release.Manifest, "\n"), nil
}

// ListReleases list releases in a specific namespace if given
func (p *Proxy) ListReleases(namespace string, releaseListLimit int) ([]AppOverview, error) {
	list, err := p.helmClient.ListReleases(
		helm.ReleaseListLimit(releaseListLimit),
		helm.ReleaseListNamespace(namespace),
		helm.ReleaseListStatuses(releaseStatuses),
	)
	if err != nil {
		return []AppOverview{}, fmt.Errorf("Unable to list helm releases: %v", err)
	}
	appList := []AppOverview{}
	if list != nil {
		for _, r := range list.Releases {
			if namespace == "" || namespace == r.Namespace {
				appList = append(appList, AppOverview{
					ReleaseName: r.Name,
					Version:     r.Chart.Metadata.Version,
					Namespace:   r.Namespace,
					Icon:        r.Chart.Metadata.Icon,
					Status:      r.Info.Status.Code.String(),
				})
			}
		}
	}
	return appList, nil
}

func lock(name string) {
	if appMutex[name] == nil {
		appMutex[name] = &sync.Mutex{}
	}
	appMutex[name].Lock()
}

func unlock(name string) {
	appMutex[name].Unlock()
}

// CreateRelease creates a tiller release
func (p *Proxy) CreateRelease(name, namespace, values string, ch *chart.Chart) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	log.Printf("Installing release %s into namespace %s", name, namespace)
	res, err := p.helmClient.InstallReleaseFromChart(
		ch,
		namespace,
		helm.ValueOverrides([]byte(values)),
		helm.ReleaseName(name),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to create the release: %v", err)
	}
	log.Printf("%s successfully installed in %s", name, namespace)
	return res.GetRelease(), nil
}

// UpdateRelease upgrades a tiller release
func (p *Proxy) UpdateRelease(name, namespace string, values string, ch *chart.Chart) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	// Check if the release already exists
	_, err := p.get(name, namespace)
	if err != nil && isNotFound(err) {
		return nil, fmt.Errorf("Release %s not found in the namespace %s. Unable to update it", name, namespace)
	} else if err != nil {
		return nil, err
	}
	log.Printf("Updating release %s", name)
	res, err := p.helmClient.UpdateReleaseFromChart(
		name,
		ch,
		helm.UpdateValueOverrides([]byte(values)),
		//helm.UpgradeForce(true), ?
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to update the release: %v", err)
	}
	return res.GetRelease(), nil
}

// GetRelease returns the info of a release
func (p *Proxy) GetRelease(name, namespace string) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	return p.get(name, namespace)
}

// DeleteRelease deletes a release
func (p *Proxy) DeleteRelease(name, namespace string) error {
	lock(name)
	defer unlock(name)
	// Validate that the release actually belongs to the namespace
	_, err := p.get(name, namespace)
	if err != nil {
		return err
	}
	_, err = p.helmClient.DeleteRelease(name, helm.DeletePurge(true))
	if err != nil {
		return fmt.Errorf("Unable to delete the release: %v", err)
	}
	return nil
}

// TillerClient for exposed funcs
type TillerClient interface {
	GetReleaseStatus(relName string) (release.Status_Code, error)
	ResolveManifest(namespace, values string, ch *chart.Chart) (string, error)
	ListReleases(namespace string, releaseListLimit int) ([]AppOverview, error)
	CreateRelease(name, namespace, values string, ch *chart.Chart) (*release.Release, error)
	UpdateRelease(name, namespace string, values string, ch *chart.Chart) (*release.Release, error)
	GetRelease(name, namespace string) (*release.Release, error)
	DeleteRelease(name, namespace string) error
}
