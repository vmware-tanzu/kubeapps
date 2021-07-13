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
	"encoding/json"
	"fmt"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	log "k8s.io/klog/v2"
)

// calling this file utils.go until I can come up with better name or organize code differently
func prettyPrintObject(o runtime.Object) string {
	prettyBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	return string(prettyBytes)
}

func prettyPrintMap(m map[string]interface{}) string {
	prettyBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", m)
	}
	return string(prettyBytes)
}

func indexOneRepo(unstructuredRepo map[string]interface{}) ([]chart.Chart, error) {
	startTime := time.Now()

	repo, err := newPackageRepository(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	ready, err := isRepoReady(unstructuredRepo)
	if err != nil || !ready {
		return nil, status.Errorf(codes.Internal,
			"cannot index repository [%s] because it is not in 'Ready' state:%v\n%s",
			repo.Name,
			err,
			prettyPrintMap(unstructuredRepo))
	}

	indexUrl, found, err := unstructured.NestedString(unstructuredRepo, "status", "url")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal,
			"expected field status.url not found on HelmRepository [%s]: %v:\n%s",
			repo.Name,
			err,
			prettyPrintMap(unstructuredRepo))
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

	// this is potentially a very expensive operation for large repos like 'bitnami'
	charts, err := helm.ChartsFromIndex(bytes, modelRepo, true)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	log.Infof("Indexed [%d] packages in repository [%s] in [%d] ms", len(charts), repo.Name, duration.Milliseconds())
	return charts, nil
}

func newPackageRepository(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field metadata.name not found on HelmRepository: %v:\n%s", err, prettyPrintMap(unstructuredRepo))
	}
	namespace, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"field metadata.namespace not found on HelmRepository: %v:\n%s", err, prettyPrintMap(unstructuredRepo))
	}
	url, found, err := unstructured.NestedString(unstructuredRepo, "spec", "url")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field spec.url not found on HelmRepository: %v:\n%s", err, prettyPrintMap(unstructuredRepo))
	}
	return &v1alpha1.PackageRepository{
		Name:      name,
		Namespace: namespace,
		Url:       url,
	}, nil
}

// isValidChart returns true if the chart model passed defines a value
// for each required field described at the Helm website:
// https://helm.sh/docs/topics/charts/#the-chartyaml-file
// together with required fields for our model.
func isValidChart(chart *models.Chart) (bool, error) {
	if chart.Name == "" {
		return false, status.Errorf(codes.Internal, "required field .Name not found on helm chart: %v", chart)
	}
	if chart.ID == "" {
		return false, status.Errorf(codes.Internal, "required field .ID not found on helm chart: %v", chart)
	}
	if chart.Repo == nil {
		return false, status.Errorf(codes.Internal, "required field .Repo not found on helm chart: %v", chart)
	}
	if chart.ChartVersions == nil || len(chart.ChartVersions) == 0 {
		return false, status.Errorf(codes.Internal, "required field .chart.ChartVersions[0] not found on helm chart: %v", chart)
	} else {
		for _, chartVersion := range chart.ChartVersions {
			if chartVersion.Version == "" {
				return false, status.Errorf(codes.Internal, "required field .ChartVersions[i].Version not found on helm chart: %v", chart)
			}
		}
	}
	if chart.Maintainers != nil || len(chart.ChartVersions) != 0 {
		for _, maintainer := range chart.Maintainers {
			if maintainer.Name == "" {
				return false, status.Errorf(codes.Internal, "required field .Maintainers[i].Name not found on helm chart: %v", chart)
			}
		}
	}
	return true, nil
}

// AvailablePackageSummaryFromChart builds an AvailablePackageSummary from a Chart
func AvailablePackageSummaryFromChart(chart *models.Chart) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	isValid, err := isValidChart(chart)
	if !isValid || err != nil {
		return nil, status.Errorf(codes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.DisplayName = chart.Name
	pkg.IconUrl = chart.Icon
	pkg.ShortDescription = chart.Description

	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     GetPluginDetail(),
	}
	pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}

	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.LatestPkgVersion = chart.ChartVersions[0].Version
	}

	return pkg, nil
}

// implements plug-in specific cache-related functionality
// onAddOrModifyRepo essentially tells the cache what to store for a given key
func onAddOrModifyRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	ready, err := isRepoReady(unstructuredRepo)
	if err != nil {
		return nil, false, err
	}

	if ready {
		charts, err := indexOneRepo(unstructuredRepo)
		if err != nil {
			return nil, false, err
		}

		jsonBytes, err := json.Marshal(charts)
		if err != nil {
			return nil, false, err
		}

		return jsonBytes, true, nil
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func onGetRepo(key string, value interface{}) (interface{}, error) {
	bytes, ok := value.([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unexpected value found in cache for key [%s]: %v", key, value)
	}

	var charts []chart.Chart
	err := json.Unmarshal(bytes, &charts)
	if err != nil {
		return nil, err
	}
	return charts, nil
}

func onDeleteRepo(key string, unstructuredRepo map[string]interface{}) (bool, error) {
	return true, nil
}
