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
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	appMutex           map[string]*sync.Mutex
	allReleaseStatuses []release.Status_Code
)

func init() {
	appMutex = make(map[string]*sync.Mutex)
	// List of posible statuses obtained from:
	// https://github.com/helm/helm/blob/master/cmd/helm/list.go#L214
	allReleaseStatuses = []release.Status_Code{
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

// NewProxy creates a Proxy
func NewProxy(kubeClient kubernetes.Interface, helmClient helm.Interface) *Proxy {
	return &Proxy{
		kubeClient: kubeClient,
		helmClient: helmClient,
	}
}

// AppOverview represents the basics of a release
type AppOverview struct {
	ReleaseName   string         `json:"releaseName"`
	Version       string         `json:"version"`
	Namespace     string         `json:"namespace"`
	Icon          string         `json:"icon,omitempty"`
	Status        string         `json:"status"`
	Chart         string         `json:"chart"`
	ChartMetadata chart.Metadata `json:"chartMetadata"`
}

func (p *Proxy) getRelease(name, namespace string) (*release.Release, error) {
	release, err := p.helmClient.ReleaseContent(name)
	if err != nil {
		return nil, prettyError(err)
	}

	// We check that the release found is from the provided namespace.
	// If `namespace` is an empty string we do not do that check
	// This check check is to prevent users of for example updating releases that might be
	// in namespaces that they do not have access to.
	if namespace != "" && release.Release.Namespace != namespace {
		return nil, fmt.Errorf("Release %q not found in namespace %q", name, namespace)
	}

	return release.Release, nil
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

// ResolveManifestFromRelease returns a manifest given the release name and revision
func (p *Proxy) ResolveManifestFromRelease(releaseName string, revision int32) (string, error) {
	// We use the release returned after running a dry-run to know the components of the release
	res, err := p.helmClient.ReleaseContent(
		releaseName,
		helm.ContentReleaseVersion(revision),
	)
	if err != nil {
		return "", err
	}
	// The manifest returned has some extra new lines at the beginning
	return strings.TrimLeft(res.Release.Manifest, "\n"), nil
}

// Apply the same filtering than helm CLI
// Ref: https://github.com/helm/helm/blob/d3b69c1fc1ac62f1cc40f93fcd0cba275c0596de/cmd/helm/list.go#L173
func filterList(rels []*release.Release) []*release.Release {
	idx := map[string]int32{}

	for _, r := range rels {
		name, version := r.GetName(), r.GetVersion()
		if max, ok := idx[name]; ok {
			// check if we have a greater version already
			if max > version {
				continue
			}
		}
		idx[name] = version
	}

	uniq := make([]*release.Release, 0, len(idx))
	for _, r := range rels {
		if idx[r.GetName()] == r.GetVersion() {
			uniq = append(uniq, r)
		}
	}
	return uniq
}

// getStatuses follows the same approach than helm CLI:
// https://github.com/helm/helm/blob/8761bb009f4eb4bcbbe7d20e434047e22b2046ad/cmd/helm/list.go#L212
func getStatuses(statusQuery string) []release.Status_Code {
	if statusQuery == "" {
		// Default case
		return []release.Status_Code{
			release.Status_DEPLOYED,
			release.Status_FAILED,
		}
	} else if strings.Contains(statusQuery, "all") {
		return allReleaseStatuses
	} else {
		statuses := []release.Status_Code{}
		for _, s := range strings.Split(statusQuery, ",") {
			switch strings.ToLower(s) {
			case "deployed":
				statuses = append(statuses, release.Status_DEPLOYED)
			case "deleted":
				statuses = append(statuses, release.Status_DELETED)
			case "deleting":
				statuses = append(statuses, release.Status_DELETING)
			case "failed":
				statuses = append(statuses, release.Status_FAILED)
			case "superseded":
				statuses = append(statuses, release.Status_SUPERSEDED)
			case "pending":
				statuses = append(statuses, release.Status_PENDING_INSTALL, release.Status_PENDING_UPGRADE, release.Status_PENDING_ROLLBACK)
			default:
				log.Infof("Ignoring unrecognized status %s", s)
			}
		}
		return statuses
	}
}

// ListReleases list releases in a specific namespace if given
func (p *Proxy) ListReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error) {
	list, err := p.helmClient.ListReleases(
		helm.ReleaseListLimit(releaseListLimit),
		helm.ReleaseListNamespace(namespace),
		helm.ReleaseListStatuses(getStatuses(status)),
	)
	if err != nil {
		return []AppOverview{}, fmt.Errorf("Unable to list helm releases: %v", err)
	}
	appList := []AppOverview{}
	if list != nil {
		filteredReleases := filterList(list.GetReleases())
		for _, r := range filteredReleases {
			if namespace == "" || namespace == r.Namespace {
				appList = append(appList, AppOverview{
					ReleaseName:   r.Name,
					Version:       r.Chart.Metadata.Version,
					Namespace:     r.Namespace,
					Icon:          r.Chart.Metadata.Icon,
					Status:        r.Info.Status.Code.String(),
					Chart:         r.Chart.Metadata.Name,
					ChartMetadata: *r.Chart.Metadata,
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
	_, err := p.getRelease(name, namespace)
	if err != nil {
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

// RollbackRelease rolls back to a specific revision
func (p *Proxy) RollbackRelease(name, namespace string, revision int32) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	// Check if the release already exists
	_, err := p.getRelease(name, namespace)
	if err != nil {
		return nil, err
	}
	res, err := p.helmClient.RollbackRelease(
		name,
		helm.RollbackVersion(revision),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to rollback the release: %v", err)
	}
	return res.GetRelease(), nil
}

// GetRelease returns the info of a release
func (p *Proxy) GetRelease(name, namespace string) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	return p.getRelease(name, namespace)
}

// DeleteRelease deletes a release
func (p *Proxy) DeleteRelease(name, namespace string, purge bool) error {
	lock(name)
	defer unlock(name)
	// Validate that the release actually belongs to the namespace
	_, err := p.getRelease(name, namespace)
	if err != nil {
		return err
	}
	_, err = p.helmClient.DeleteRelease(name, helm.DeletePurge(purge))
	if err != nil {
		return fmt.Errorf("Unable to delete the release: %v", err)
	}
	return nil
}

// extracted from https://github.com/helm/helm/blob/master/cmd/helm/helm.go#L227
// prettyError unwraps or rewrites certain errors to make them more user-friendly.
func prettyError(err error) error {
	// Add this check can prevent the object creation if err is nil.
	if err == nil {
		return nil
	}
	// If it's grpc's error, make it more user-friendly.
	if s, ok := status.FromError(err); ok {
		return fmt.Errorf(s.Message())
	}
	// Else return the original error.
	return err
}

// TillerClient for exposed funcs
type TillerClient interface {
	GetReleaseStatus(relName string) (release.Status_Code, error)
	ResolveManifest(namespace, values string, ch *chart.Chart) (string, error)
	ResolveManifestFromRelease(releaseName string, revision int32) (string, error)
	ListReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error)
	CreateRelease(name, namespace, values string, ch *chart.Chart) (*release.Release, error)
	UpdateRelease(name, namespace string, values string, ch *chart.Chart) (*release.Release, error)
	RollbackRelease(name, namespace string, revision int32) (*release.Release, error)
	GetRelease(name, namespace string) (*release.Release, error)
	DeleteRelease(name, namespace string, purge bool) error
}
