// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httphandler

import (
	"github.com/gorilla/mux"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
)

// SetupDefaultRoutes enables call-sites to use the backend api's default routes with minimal setup.
func SetupDefaultRoutes(_ *mux.Router, _ int, _ float32, _ kube.ClustersConfig) error {
	return nil
}
