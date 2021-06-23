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
	"context"
	"fmt"
	"os"

	"github.com/kubeapps/common/datastore"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

// Server implements the helm packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedHelmPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter func(context.Context) (dynamic.Interface, error)
	manager      assetManager
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(clientGetter func(context.Context) (dynamic.Interface, error)) *Server {
	var kubeappsNamespace = os.Getenv("POD_NAMESPACE")
	var ASSET_SYNCER_DB_URL = os.Getenv("ASSET_SYNCER_DB_URL")
	var ASSET_SYNCER_DB_NAME = os.Getenv("ASSET_SYNCER_DB_NAME")
	var ASSET_SYNCER_DB_USERNAME = os.Getenv("ASSET_SYNCER_DB_USERNAME")
	var ASSET_SYNCER_DB_USERPASSWORD = os.Getenv("ASSET_SYNCER_DB_USERPASSWORD")

	var dbConfig = datastore.Config{URL: ASSET_SYNCER_DB_URL, Database: ASSET_SYNCER_DB_NAME, Username: ASSET_SYNCER_DB_USERNAME, Password: ASSET_SYNCER_DB_USERPASSWORD}

	manager, err := newPGManager(dbConfig, kubeappsNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}

	return &Server{
		clientGetter: clientGetter,
		manager:      manager,
	}
}

// GetClient ensures a client getter is available and uses it to return the client.
func (s *Server) GetClient(ctx context.Context) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	client, err := s.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return client, nil
}

// GetManager ensures a manager is available and uses it to return the client.
func (s *Server) GetManager() (assetManager, error) {
	if s.manager == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with manager")
	}
	manager := s.manager
	return manager, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {

	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+helm GetAvailablePackageSummaries %s", contextMsg)

	namespace := ""
	if request.Context != nil {
		if request.Context.Cluster != "" {
			return nil, status.Errorf(codes.Unimplemented, "Not supported yet: request.Context.Cluster: [%v]", request.Context.Cluster)
		}
		if request.Context.Namespace != "" {
			namespace = request.Context.Namespace
		}
	}

	// TODO: add more filters in the context?
	cq := ChartQuery{
		namespace: namespace,
	}

	// We are not returning paginated results here
	charts, _, err := s.manager.getPaginatedChartListWithFilters(cq, 1, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve charts: %v", err)
	}

	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg, err := AvailablePackageSummaryFromChart(chart)
		if err != nil {
			log.Fatalf("%s", err)
			return nil, err
		}
		responsePackages = append(responsePackages, pkg)
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: responsePackages,
	}, nil
}

func AvailablePackageSummaryFromChart(chart *models.Chart) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	if chart.Name == "" {
		return nil, status.Errorf(codes.Internal, "required field .Name not found on helm package: %v", chart)
	}
	pkg.DisplayName = chart.Name

	if chart.ChartVersions == nil || len(chart.ChartVersions) == 0 || chart.ChartVersions[0].Version == "" {
		return nil, status.Errorf(codes.Internal, "required field .ChartVersions[0].Version not found on helm package: %v", chart)
	}
	pkg.LatestVersion = chart.ChartVersions[0].Version

	if chart.Icon == "" {
		return nil, status.Errorf(codes.Internal, "required field .Icon not found on helm package: %v", chart)
	}
	pkg.IconUrl = chart.Icon

	if chart.Description == "" {
		return nil, status.Errorf(codes.Internal, "required field .Description not found on helm package: %v", chart)
	}
	pkg.ShortDescription = chart.Description

	if chart.ID == "" {
		return nil, status.Errorf(codes.Internal, "required field .ID not found on helm package: %v", chart)
	}
	if chart.Repo == nil || chart.Repo.Namespace == "" {
		return nil, status.Errorf(codes.Internal, "required field .Repo.Namespace not found on helm package: %v", chart)
	}
	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Context:    &corev1.Context{Namespace: chart.Repo.Namespace},
		Identifier: chart.ID,
	}

	return pkg, nil
}
