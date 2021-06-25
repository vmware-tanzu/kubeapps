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
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	clientGetter             server.KubernetesClientGetter
	globalPackagingNamespace string
	manager                  utils.AssetManager
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(clientGetter server.KubernetesClientGetter) *Server {
	var kubeappsNamespace = os.Getenv("POD_NAMESPACE")
	var ASSET_SYNCER_DB_URL = os.Getenv("ASSET_SYNCER_DB_URL")
	var ASSET_SYNCER_DB_NAME = os.Getenv("ASSET_SYNCER_DB_NAME")
	var ASSET_SYNCER_DB_USERNAME = os.Getenv("ASSET_SYNCER_DB_USERNAME")
	var ASSET_SYNCER_DB_USERPASSWORD = os.Getenv("ASSET_SYNCER_DB_USERPASSWORD")

	var dbConfig = datastore.Config{URL: ASSET_SYNCER_DB_URL, Database: ASSET_SYNCER_DB_NAME, Username: ASSET_SYNCER_DB_USERNAME, Password: ASSET_SYNCER_DB_USERPASSWORD}

	manager, err := utils.NewPGManager(dbConfig, kubeappsNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}

	return &Server{
		clientGetter:             clientGetter,
		manager:                  manager,
		globalPackagingNamespace: kubeappsNamespace,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}

// GetManager ensures a manager is available and returns it.
func (s *Server) GetManager() (utils.AssetManager, error) {
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

	// Check the request context (namespace and cluster)
	namespace := ""
	if request.Context != nil {
		if request.Context.Cluster != "" {
			return nil, status.Errorf(codes.Unimplemented, "Not supported yet: request.Context.Cluster: [%v]", request.Context.Cluster)
		}
		namespace = request.Context.Namespace
	}
	// Check the requested namespace: if any, return "everything a user can read";
	// otherwise, first check if the user can access the requested ns
	if namespace == "" {
		// TODO(agamez): not including a namespace means that it returns everything a user can read
		return nil, status.Errorf(codes.Unimplemented, "Not supported yet: not including a namespace means that it returns everything a user can read")
	}
	// After requesting a specific namespace, we have to ensure the user can actually access to it
	// If checking the global namespace, allow access always
	hasAccess := namespace != s.globalPackagingNamespace
	if !hasAccess {
		var err error
		// If checking another namespace, check if the user has access (ie, "get secrets in this ns")
		hasAccess, err = s.hasAccessToNamespace(ctx, namespace)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to check if the user has access to the namespace: %s", err)
		}
		if !hasAccess {
			// If the user has not access, return a unauthenticated response, otherwise, continue
			return nil, status.Errorf(codes.Unauthenticated, "The current user has no access to the namespace %q", namespace)
		}
	}

	// Create the initial chart query with the namespace
	cq := utils.ChartQuery{
		Namespace: namespace,
	}

	// Add any other filter if a FilterOptions is passed
	if request.FilterOptions != nil {
		cq.Categories = request.FilterOptions.Categories
		cq.SearchQuery = request.FilterOptions.Query
		cq.Repos = request.FilterOptions.Repositories
		cq.Version = request.FilterOptions.Version
		cq.AppVersion = request.FilterOptions.AppVersion
	}

	// TODO: We are not yet returning paginated results here
	charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, 1, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve charts: %v", err)
	}

	// Convert the charts response into a GetAvailablePackageSummariesResponse
	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg, err := AvailablePackageSummaryFromChart(chart)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to parse chart to an AvailablePackageSummary: %v", err)
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

func (s *Server) hasAccessToNamespace(ctx context.Context, namespace string) (bool, error) {
	client, _, err := s.GetClients(ctx)
	if err != nil {
		return false, err
	}

	res, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Group:     "",
				Resource:  "secrets",
				Verb:      "get",
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return res.Status.Allowed, nil
}
