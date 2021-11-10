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
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
)

// Currently just a stub unimplemented server. More to come in following PRs.
type Server struct {
	v1alpha1.UnimplementedResourcesServiceServer
}

func NewServer(configGetter server.KubernetesConfigGetter) *Server {
	return &Server{}
}
