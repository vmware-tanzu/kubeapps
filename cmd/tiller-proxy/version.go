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

package main

import (
	"fmt"
)

var (
	version          = "devel"
	userAgentComment string
)

// Returns the user agent to be used during calls to the chart repositories
// Examples:
// tiller-proxy/devel
// chart-repo/1.0
// tiller-proxy/1.0 (monocular v1.0-beta4)
// More info here https://github.com/kubeapps/kubeapps/issues/767#issuecomment-436835938
func userAgent() string {
	ua := "tiller-proxy/" + version
	if userAgentComment != "" {
		ua = fmt.Sprintf("%s (%s)", ua, userAgentComment)
	}
	return ua
}
