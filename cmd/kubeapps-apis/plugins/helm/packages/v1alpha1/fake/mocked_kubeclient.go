package fake

import (
	"io"
	"time"

	v1 "k8s.io/api/core/v1"

	kube "helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
)

// DelayedKubeClient implements KubeClient for testing purposes.
// Sleeps for "Time" before returning each operation delegated to "PrintingKubeClient"
type DelayedKubeClient struct {
	kubefake.PrintingKubeClient
	Time time.Duration
}

func (f *DelayedKubeClient) Create(resources kube.ResourceList) (*kube.Result, error) {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.Create(resources)
}

func (f *DelayedKubeClient) Wait(resources kube.ResourceList, d time.Duration) error {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.Wait(resources, d)
}

func (f *DelayedKubeClient) WaitWithJobs(resources kube.ResourceList, d time.Duration) error {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.WaitWithJobs(resources, d)
}

func (f *DelayedKubeClient) WaitForDelete(resources kube.ResourceList, d time.Duration) error {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.WaitForDelete(resources, d)
}

func (f *DelayedKubeClient) Delete(resources kube.ResourceList) (*kube.Result, []error) {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.Delete(resources)
}

func (f *DelayedKubeClient) WatchUntilReady(resources kube.ResourceList, d time.Duration) error {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.WatchUntilReady(resources, d)
}

func (f *DelayedKubeClient) Update(r, modified kube.ResourceList, ignoreMe bool) (*kube.Result, error) {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.Update(r, modified, ignoreMe)
}

func (f *DelayedKubeClient) Build(r io.Reader, _ bool) (kube.ResourceList, error) {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.Build(r, false)
}

func (f *DelayedKubeClient) WaitAndGetCompletedPodPhase(s string, d time.Duration) (v1.PodPhase, error) {
	time.Sleep(f.Time)
	return f.PrintingKubeClient.WaitAndGetCompletedPodPhase(s, d)
}
