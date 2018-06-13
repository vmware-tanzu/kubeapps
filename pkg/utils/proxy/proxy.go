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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"

	chartUtils "github.com/kubeapps/kubeapps/pkg/utils/chart"
)

const (
	defaultNamespace      = metav1.NamespaceSystem
	defaultRepoURL        = "https://kubernetes-charts.storage.googleapis.com"
	defaultTimeoutSeconds = 180
)

var (
	appMutex map[string]*sync.Mutex
)

func init() {
	appMutex = make(map[string]*sync.Mutex)
}

// Proxy contains all the elements to contact Tiller and the K8s API
type Proxy struct {
	kubeClient kubernetes.Interface
	helmClient helm.Interface
	netClient  *chartUtils.HTTPClient
	loadChart  chartUtils.LoadChart
}

type helmRelease struct {
	// RepoURL is the URL of the repository. Defaults to stable repo.
	RepoURL string `json:"repoUrl,omitempty"`
	// ChartName is the name of the chart within the repo.
	ChartName string `json:"chartName"`
	// ReleaseName is the Name of the release given to Tiller.
	ReleaseName string `json:"releaseName"`
	// Version is the chart version.
	Version string `json:"version"`
	// Auth is the authentication.
	Auth helmReleaseAuth `json:"auth,omitempty"`
	// Values is a string containing (unparsed) YAML values.
	Values string `json:"values,omitempty"`
}

type helmReleaseAuth struct {
	// Header is header based Authorization
	Header *helmReleaseAuthHeader `json:"header,omitempty"`
}

type helmReleaseAuthHeader struct {
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

func isNotFound(err error) bool {
	// Ideally this would be `grpc.Code(err) == codes.NotFound`,
	// but it seems helm doesn't return grpc codes
	return strings.Contains(grpc.ErrorDesc(err), "not found")
}

// NewProxy creates a Proxy
func NewProxy(kubeClient kubernetes.Interface, helmClient helm.Interface, netClient chartUtils.HTTPClient, loadChart chartUtils.LoadChart) *Proxy {
	return &Proxy{
		kubeClient: kubeClient,
		helmClient: helmClient,
		netClient:  &netClient,
		loadChart:  loadChart,
	}
}

// AppOverview represents the basics of a release
type AppOverview struct {
	ReleaseName string `json:"releaseName"`
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
}

func (p *Proxy) getChart(rel *helmRelease) (*chart.Chart, error) {
	repoURL := rel.RepoURL
	if repoURL == "" {
		// FIXME: Make configurable
		repoURL = defaultRepoURL
	}
	repoURL = strings.TrimSuffix(strings.TrimSpace(repoURL), "/") + "/index.yaml"

	authHeader := ""
	if rel.Auth.Header != nil {
		namespace := os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			namespace = defaultNamespace
		}

		secret, err := p.kubeClient.Core().Secrets(namespace).Get(rel.Auth.Header.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		authHeader = string(secret.Data[rel.Auth.Header.SecretKeyRef.Key])
	}

	log.Printf("Downloading repo %s index...", repoURL)
	repoIndex, err := chartUtils.FetchRepoIndex(p.netClient, repoURL, authHeader)
	if err != nil {
		return nil, err
	}

	chartURL, err := chartUtils.FindChartInRepoIndex(repoIndex, repoURL, rel.ChartName, rel.Version)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading %s ...", chartURL)
	chartRequested, err := chartUtils.FetchChart(p.netClient, chartURL, authHeader, p.loadChart)
	if err != nil {
		return nil, err
	}
	return chartRequested, nil
}

func (p *Proxy) get(name, namespace string) (*release.Release, error) {
	list, err := p.helmClient.ListReleases()
	if err != nil {
		return nil, fmt.Errorf("Unable to list helm releases: %v", err)
	}
	var rel *release.Release
	for _, r := range list.Releases {
		if (namespace == "" || namespace == r.Namespace) && r.Name == name {
			rel = r
			break
		}
	}
	if rel == nil {
		return nil, fmt.Errorf("Release %s not found in namespace %s", name, namespace)
	}
	return rel, nil
}

func (p *Proxy) deploy(name, namespace, values string, ch *chart.Chart) (*release.Release, error) {
	log.Printf("Installing release %s into namespace %s", name, namespace)
	res, err := p.helmClient.InstallReleaseFromChart(
		ch,
		namespace,
		helm.ValueOverrides([]byte(values)),
		helm.ReleaseName(name),
	)
	if err != nil {
		return nil, err
	}
	return res.GetRelease(), nil
}

func (p *Proxy) update(name, namespace, values string, ch *chart.Chart) (*release.Release, error) {
	log.Printf("Updating release %s", name)
	res, err := p.helmClient.UpdateReleaseFromChart(
		name,
		ch,
		helm.UpdateValueOverrides([]byte(values)),
		//helm.UpgradeForce(true), ?
	)
	if err != nil {
		return nil, err
	}
	return res.GetRelease(), nil
}

// LogReleaseStatus prints the status of the given release if exists
func (p *Proxy) LogReleaseStatus(relName string) {
	status, err := p.helmClient.ReleaseStatus(relName)
	if err == nil {
		if status.Info != nil && status.Info.Status != nil {
			log.Printf("Release status: %s", status.Info.Status.Code)
		}
	} else {
		log.Printf("Unable to fetch release status for %s: %v", relName, err)
	}
}

// ListReleases list releases in a specific namespace if given
func (p *Proxy) ListReleases(namespace string) ([]AppOverview, error) {
	list, err := p.helmClient.ListReleases()
	if err != nil {
		return []AppOverview{}, fmt.Errorf("Unable to list helm releases: %v", err)
	}
	appList := []AppOverview{}
	for _, r := range list.Releases {
		if namespace == "" || namespace == r.Namespace {
			appList = append(appList, AppOverview{r.Name, r.Chart.Metadata.Version, r.Namespace})
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
func (p *Proxy) CreateRelease(namespace string, rawRelease []byte) (*release.Release, error) {
	hrelease := &helmRelease{}
	err := json.Unmarshal(rawRelease, hrelease)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request body: %v", err)
	}
	lock(hrelease.ReleaseName)
	defer unlock(hrelease.ReleaseName)
	ch, err := p.getChart(hrelease)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve chart info: %v", err)
	}
	if err != nil {
		return nil, err
	}
	rel, err := p.deploy(hrelease.ReleaseName, namespace, hrelease.Values, ch)
	if err != nil {
		return nil, fmt.Errorf("Unable create release: %v", err)
	}
	return rel, nil
}

// UpdateRelease upgrades a tiller release
func (p *Proxy) UpdateRelease(name, namespace string, rawRelease []byte) (*release.Release, error) {
	lock(name)
	defer unlock(name)
	var rel *release.Release
	// Check if the release already exists
	_, err := p.get(name, namespace)
	if err != nil && isNotFound(err) {
		return nil, fmt.Errorf("Release %s not found in the namespace %s. Unable to update it", name, namespace)
	} else if err != nil {
		return nil, err
	} else {
		hrelease := &helmRelease{}
		err := json.Unmarshal(rawRelease, hrelease)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse request body: %v", err)
		}
		ch, err := p.getChart(hrelease)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve chart info: %v", err)
		}
		if err != nil {
			return nil, err
		}
		rel, err = p.update(hrelease.ReleaseName, namespace, hrelease.Values, ch)
		if err != nil {
			return nil, fmt.Errorf("Unable to update release: %v", err)
		}
	}
	return rel, nil
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
		return fmt.Errorf("Unable to delete release: %v", err)
	}
	return nil
}
