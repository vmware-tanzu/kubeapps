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

package fake

import (
	"fmt"
	"strings"

	"github.com/kubeapps/kubeapps/pkg/proxy"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type FakeProxy struct {
	Releases []release.Release
}

func (f *FakeProxy) GetReleaseStatus(relName string) (release.Status_Code, error) {
	return release.Status_DEPLOYED, nil
}

func (f *FakeProxy) ResolveManifest(namespace, values string, ch *chart.Chart) (string, error) {
	return "", nil
}

func (f *FakeProxy) ResolveManifestFromRelease(releaseName string, revision int32) (string, error) {
	return "", nil
}

func (f *FakeProxy) ListReleases(namespace string, releaseListLimit int, status string) ([]proxy.AppOverview, error) {
	res := []proxy.AppOverview{}
	for _, r := range f.Releases {
		relStatus := "DEPLOYED" // Default
		if r.Info != nil {
			relStatus = r.Info.Status.Code.String()
		}
		if (namespace == "" || namespace == r.Namespace) &&
			len(res) <= releaseListLimit &&
			(r.Info == nil || status == strings.ToLower(relStatus)) {
			res = append(res, proxy.AppOverview{
				ReleaseName: r.Name,
				Version:     "",
				Namespace:   r.Namespace,
				Icon:        "",
				Status:      relStatus,
			})
		}
	}
	return res, nil
}

func (f *FakeProxy) CreateRelease(name, namespace, values string, ch *chart.Chart) (*release.Release, error) {
	for _, r := range f.Releases {
		if r.Name == name {
			return nil, fmt.Errorf("Release already exists")
		}
	}
	r := release.Release{
		Name:      name,
		Namespace: namespace,
	}
	f.Releases = append(f.Releases, r)
	return &r, nil
}

func (f *FakeProxy) UpdateRelease(name, namespace string, values string, ch *chart.Chart) (*release.Release, error) {
	for _, r := range f.Releases {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("Release %s not found", name)
}

func (f *FakeProxy) RollbackRelease(name, namespace string, revision int32) (*release.Release, error) {
	for _, r := range f.Releases {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("Release %s not found", name)
}

func (f *FakeProxy) GetRelease(name, namespace string) (*release.Release, error) {
	for _, r := range f.Releases {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("Release %s not found", name)
}

func (f *FakeProxy) DeleteRelease(name, namespace string, purge bool) error {
	for i, r := range f.Releases {
		if r.Name == name {
			if purge {
				f.Releases[i] = f.Releases[len(f.Releases)-1]
				f.Releases = f.Releases[:len(f.Releases)-1]
			} else {
				r.Info = &release.Info{
					Status: &release.Status{
						Code: release.Status_DELETED,
					},
				}
				f.Releases[i] = r
			}
			return nil
		}
	}
	return fmt.Errorf("Release %s not found", name)
}
