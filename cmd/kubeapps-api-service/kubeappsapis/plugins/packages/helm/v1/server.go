package v1

import (
	v1 "github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/kubeappsapis/core/packages/v1"
)

// Server implements the helm package repositories v1 interface.
type Server struct {
	UnimplementedPackagesServiceServer
}

func (s *Server) GetInstalledPackages(request *v1.GetInstalledPackagesRequest, stream PackagesService_GetInstalledPackagesServer) error {
	installedPackages := []*v1.InstalledPackageSummary{
		{
			Name:      "Apache",
			Namespace: "user1",
			Version:   "6.8.0",
			IconUrl:   "http://example.com/apache.jpg",
		},
		{
			Name:      "nginx",
			Namespace: "user1",
			Version:   "3.4.0",
			IconUrl:   "http://example.com/nginx.jpg",
		},
	}
	for _, pkg := range installedPackages {
		stream.Send(pkg)
	}
	return nil
}
