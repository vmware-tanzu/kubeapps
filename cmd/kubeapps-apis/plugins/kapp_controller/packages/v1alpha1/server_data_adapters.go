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
	"fmt"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
)

func (s *Server) getAvailablePackageSummary(pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, cluster string) (*corev1.AvailablePackageSummary, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgMetadata.Name]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name)
	}

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.
	iconUrl := ""
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconUrl = fmt.Sprintf("data:image/svg+xml;base64,%s", pkgMetadata.Spec.IconSVGBase64)
	}

	availablePackageSummary := &corev1.AvailablePackageSummary{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgMetadata.Namespace,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgMetadata.Name,
		},
		Name: pkgMetadata.Spec.DisplayName,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
		},
		IconUrl:          iconUrl,
		DisplayName:      pkgMetadata.Spec.DisplayName,
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Categories:       pkgMetadata.Spec.Categories,
	}

	return availablePackageSummary, nil
}
