package fake

import (
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	fakeChart "github.com/kubeapps/kubeapps/pkg/chart/fake"
)

// ClientResolver implements ResolverFactory
type ClientResolver struct{}

// New for ClientResolver
func (c *ClientResolver) New(repoType, userAgent string) chartUtils.Resolver {
	return &fakeChart.Client{}
}
