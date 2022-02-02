// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/kubeapps/kubeapps/pkg/helm"
	helmregistry "helm.sh/helm/v3/pkg/registry"
)

// FakeOCIClient embeds *helmregistry.Clien, but overrides the Pull method
type FakeOCIClient struct {
	*helmregistry.Client

	ExpectedName string
	Content      map[string]*bytes.Buffer
	Checksum     string
	Err          error
}

// OCIClientFactoryMock is the factory used for creating FakeOCIClients
type OCIClientFactoryMock struct {
	ExpectedName string
	Content      map[string]*bytes.Buffer
	Checksum     string
	Err          error
}

// BuildOCIClient just passes the fileds from the factory to the FakeOCIClient
func (cf OCIClientFactoryMock) BuildOCIClient(resolver remotes.Resolver) (helm.IOCIClient, error) {
	client, err := helmregistry.NewClient()
	return FakeOCIClient{
		Client:       client,
		ExpectedName: cf.ExpectedName,
		Content:      cf.Content,
		Checksum:     cf.Checksum,
		Err:          cf.Err,
	}, err
}

// Pull is the fake implementation of the helmregistry.Client.Pull method
func (f FakeOCIClient) Pull(ref string, options ...helmregistry.PullOption) (*helmregistry.PullResult, error) {
	tag := strings.Split(ref, ":")[1]
	if f.ExpectedName != "" && f.ExpectedName != ref {
		return nil, fmt.Errorf("expecting %s got %s", f.ExpectedName, ref)
	}
	pullResult := &helmregistry.PullResult{
		Chart: &helmregistry.DescriptorPullSummaryWithMeta{
			DescriptorPullSummary: helmregistry.DescriptorPullSummary{
				Data:   f.Content[tag].Bytes(),
				Digest: f.Checksum,
			},
		},
	}
	return pullResult, f.Err
}
