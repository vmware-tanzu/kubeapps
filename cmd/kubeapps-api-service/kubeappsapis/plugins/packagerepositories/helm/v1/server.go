package v1

import (
	v1 "github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/kubeappsapis/core/packagerepositories/v1"
)

// Server implements the helm package repositories v1 interface.
type Server struct {
	UnimplementedPackageRepositoriesServiceServer
}

func (s *Server) GetAvailablePackages(request *v1.GetAvailablePackagesRequest, stream PackageRepositoriesService_GetAvailablePackagesServer) error {
	repo := &v1.PackageRepository{
		Name:      "bitnami",
		Namespace: "kubeapps",
	}
	availablePackages := []*v1.AvailablePackage{
		{
			Name:          "package-a",
			LatestVersion: "1.2.0",
			Repository:    repo,
			IconUrl:       "http://example.com/package-a.jpg",
		},
		{
			Name:          "package-b",
			Repository:    repo,
			LatestVersion: "1.4.0",
			IconUrl:       "http://example.com/package-b.jpg",
		},
	}
	for _, pkg := range availablePackages {
		stream.Send(pkg)
	}
	return nil
}
