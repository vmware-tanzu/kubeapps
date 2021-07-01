/*
Copyright © 2021 VMware
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
package main

import (
	"math"
	"sync"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	log "k8s.io/klog/v2"
)

// calling this file utils.go until I can come up with better name or organize code differently

const (
	// max number of concurrent workers reading repo index at the same time
	maxWorkers = 10
)

type readRepoJob struct {
	unstructuredRepo map[string]interface{}
}

type readRepoJobResult struct {
	packages []*corev1.AvailablePackageSummary
	Error    error
}

// each repo is read in a separate go routine (lightweight thread of execution)
func readPackageSummariesFromRepoList(repoItems []unstructured.Unstructured) ([]*corev1.AvailablePackageSummary, error) {
	responsePackages := []*corev1.AvailablePackageSummary{}
	var wg sync.WaitGroup
	workers := int(math.Min(float64(len(repoItems)), float64(maxWorkers)))
	readRepoJobsChannel := make(chan readRepoJob, workers)
	readRepoResultChannel := make(chan readRepoJobResult, workers)

	// Process only at most maxWorkers at a time
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			for job := range readRepoJobsChannel {
				packages, err := readPackageSummariesFromOneRepo(job.unstructuredRepo)
				readRepoResultChannel <- readRepoJobResult{packages, err}
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(readRepoResultChannel)
	}()

	go func() {
		for _, repoItem := range repoItems {
			readRepoJobsChannel <- readRepoJob{repoItem.Object}
		}
		close(readRepoJobsChannel)
	}()

	// Start receiving results
	for res := range readRepoResultChannel {
		if res.Error == nil {
			responsePackages = append(responsePackages, res.packages...)
		}
	}
	return responsePackages, nil
}

func readPackageSummariesFromOneRepo(unstructuredRepo map[string]interface{}) ([]*corev1.AvailablePackageSummary, error) {
	repo, err := newPackageRepository(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	ready, err := isRepoReady(unstructuredRepo)
	if err != nil || !ready {
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state:%v\n%v", repo.Name, err, unstructuredRepo)
		return nil, err
	}

	indexUrl, found, err := unstructured.NestedString(unstructuredRepo, "status", "url")
	if err != nil || !found {
		log.Infof("expected field status.url not found on HelmRepository [%s]: %v:\n%v", repo.Name, err, unstructuredRepo)
		return nil, err
	}

	log.Infof("Found repository: [%s], index URL: [%s]", repo.Name, indexUrl)

	// no need to provide authz, userAgent or any of the TLS details, as we are reading index.yaml file from
	// local cluster, not some remote repo.
	// e.g. http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index.yaml
	// Flux does the hard work of pulling the index file from remote repo
	// into local cluster based on secretRef associated with HelmRepository, if applicable
	bytes, err := httpclient.Get(indexUrl, httpclient.New(), map[string]string{})
	if err != nil {
		return nil, err
	}

	modelRepo := &chart.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Url,
		Type:      "helm",
	}

	// this is potentially a very expensive operation for large repos like bitnami
	charts, err := helm.ChartsFromIndex(bytes, modelRepo, true)
	if err != nil {
		return nil, err
	}

	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg := &corev1.AvailablePackageSummary{
			DisplayName:      chart.Name,
			LatestPkgVersion: chart.ChartVersions[0].Version,
			IconUrl:          chart.Icon,
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Context:    &corev1.Context{Namespace: repo.Namespace},
				Identifier: chart.ID,
			},
		}
		responsePackages = append(responsePackages, pkg)
	}
	return responsePackages, nil
}

func newPackageRepository(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field metadata.name not found on HelmRepository: %v:\n%v", err, unstructuredRepo)
	}
	namespace, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"field metadata.namespace not found on HelmRepository: %v:\n%v", err, unstructuredRepo)
	}
	url, found, err := unstructured.NestedString(unstructuredRepo, "spec", "url")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field spec.url not found on HelmRepository: %v:\n%v", err, unstructuredRepo)
	}
	return &v1alpha1.PackageRepository{
		Name:      name,
		Namespace: namespace,
		Url:       url,
	}, nil
}
