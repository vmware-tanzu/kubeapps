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

package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kubeapps/kubeapps/pkg/kube"
	yamlUtils "github.com/kubeapps/kubeapps/pkg/yaml"
	authorizationapi "k8s.io/api/authorization/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	discovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

type resource struct {
	APIVersion string
	Kind       string
	Namespace  string
}

type k8sAuthInterface interface {
	GetResourceList(groupVersion string) (*metav1.APIResourceList, error)
	CanI(verb, group, resource, namespace string) (bool, error)
}

type k8sAuth struct {
	AuthCli      authorizationv1.AuthorizationV1Interface
	DiscoveryCli discovery.DiscoveryInterface
}

func (u k8sAuth) GetResourceList(groupVersion string) (*metav1.APIResourceList, error) {
	return u.DiscoveryCli.ServerResourcesForGroupVersion(groupVersion)
}

func (u k8sAuth) CanI(verb, group, resource, namespace string) (bool, error) {
	attr := &authorizationapi.ResourceAttributes{
		Group:     group,
		Resource:  resource,
		Verb:      verb,
		Namespace: namespace,
	}
	res, err := u.AuthCli.SelfSubjectAccessReviews().Create(context.TODO(), &authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: attr,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return res.Status.Allowed, nil
}

// UserAuth contains information to check user permissions
type UserAuth struct {
	k8sAuth k8sAuthInterface
}

// Action represents a specific set of verbs against a resource
type Action struct {
	APIVersion  string   `json:"apiGroup"`
	Resource    string   `json:"resource"`
	Namespace   string   `json:"namespace"`
	ClusterWide bool     `json:"clusterWide"`
	Verbs       []string `json:"verbs"`
}

// Checker for the exported funcs
type Checker interface {
	ValidateForNamespace(namespace string) (bool, error)
	GetForbiddenActions(namespace, action, manifest string) ([]Action, error)
}

// NewAuth creates an auth agent
func NewAuth(token, clusterName string, clustersConfig kube.ClustersConfig) (*UserAuth, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config, err = kube.NewClusterConfig(config, token, clusterName, clustersConfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	authCli := kubeClient.AuthorizationV1()
	discoveryCli := kubeClient.Discovery()
	k8sAuthCli := k8sAuth{
		AuthCli:      authCli,
		DiscoveryCli: discoveryCli,
	}

	return &UserAuth{k8sAuthCli}, nil
}

// ValidateForNamespace checks if the user can access secrets in the given
// namespace, as a check of whether they can view the namespace.
func (u *UserAuth) ValidateForNamespace(namespace string) (bool, error) {
	return u.k8sAuth.CanI("get", "", "secrets", namespace)
}

type resourceInfo struct {
	Name       string
	Namespaced bool
}

func (u *UserAuth) resolve(groupVersion, kind string) (resourceInfo, error) {
	resourceList, err := u.k8sAuth.GetResourceList(groupVersion)
	if err != nil {
		return resourceInfo{}, err
	}
	for _, r := range resourceList.APIResources {
		if r.Kind == kind {
			return resourceInfo{r.Name, r.Namespaced}, nil
		}
	}
	return resourceInfo{}, fmt.Errorf("Unable to find the kind %s in the resource group %s", kind, groupVersion)
}

func (u *UserAuth) getResourcesToCheck(namespace, manifest string) ([]resource, error) {
	objs, err := yamlUtils.ParseObjects(manifest)
	if err != nil {
		return []resource{}, err
	}
	resourcesToCheck := map[string]*resource{}
	result := []resource{}
	for _, obj := range objs {
		// Object can specify a different namespace, if not use the default one
		ns := obj.GetNamespace()
		if ns == "" {
			ns = namespace
		}
		resourceToCheck := fmt.Sprintf("%s/%s/%s", ns, obj.GetAPIVersion(), obj.GetKind())
		if resourcesToCheck[resourceToCheck] == nil {
			r := resource{obj.GetAPIVersion(), obj.GetKind(), ns}
			resourcesToCheck[resourceToCheck] = &r
			result = append(result, r)
		}
	}
	return result, nil
}

func (u *UserAuth) isAllowed(verb string, itemsToCheck []resource) ([]Action, error) {
	rejectedActions := []Action{}
	for _, i := range itemsToCheck {
		rInfo, err := u.resolve(i.APIVersion, i.Kind)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				// The resource version/kind is not registered in the k8s API so
				// we assume it's a CRD that is going to be created with the chart
				// In any case, if a chart tries to install a resource that doesn't
				// exist it's fine to ignore it here since the installation will fail
				continue
			}
			return []Action{}, err
		}
		group := i.APIVersion
		if group == "v1" {
			// The group should be empty for the core API group
			group = ""
		}
		allowed, err := u.k8sAuth.CanI(verb, group, rInfo.Name, i.Namespace)
		if err != nil {
			return []Action{}, err
		}
		// If the "group" is versioned the user may be able to have access to any
		// version of the group but the above call may return "false"
		if !allowed && strings.Contains(group, "/") {
			groupID := strings.Split(group, "/")[0]
			allowed, err = u.k8sAuth.CanI(verb, groupID, rInfo.Name, i.Namespace)
			if err != nil {
				return []Action{}, err
			}
		}
		if !allowed {
			rejectedAction := Action{
				APIVersion:  i.APIVersion,
				Resource:    rInfo.Name,
				Verbs:       []string{verb},
				ClusterWide: !rInfo.Namespaced,
			}
			if rInfo.Namespaced {
				rejectedAction.Namespace = i.Namespace
			}
			rejectedActions = append(rejectedActions, rejectedAction)
		}
	}
	return rejectedActions, nil
}

func uniqVerbs(current []string, new []string) []string {
	resMap := map[string]bool{}
	for _, v := range current {
		if !resMap[v] {
			resMap[v] = true
		}
	}
	res := append([]string{}, current...)
	for _, v := range new {
		if !resMap[v] {
			resMap[v] = true
			res = append(res, v)
		}
	}
	return res
}

func reduceActionsByVerb(actions []Action) []Action {
	resMap := map[string]Action{}
	for _, action := range actions {
		req := fmt.Sprintf("%s/%s/%s", action.Namespace, action.APIVersion, action.Resource)
		if _, ok := resMap[req]; ok {
			// Element already exists
			resMap[req] = Action{
				APIVersion: action.APIVersion,
				Resource:   action.Resource,
				Namespace:  action.Namespace,
				Verbs:      uniqVerbs(resMap[req].Verbs, action.Verbs),
			}
		} else {
			resMap[req] = action
		}
	}
	res := []Action{}
	for _, a := range resMap {
		res = append(res, a)
	}
	return res
}

// GetForbiddenActions parses a K8s manifest and checks if the current user can do the action given
// over all the elements of the manifest. It return the list of forbidden Actions if any.
func (u *UserAuth) GetForbiddenActions(namespace, action, manifest string) ([]Action, error) {
	resources, err := u.getResourcesToCheck(namespace, manifest)
	forbiddenActions := []Action{}
	if err != nil {
		return []Action{}, err
	}
	switch action {
	case "upgrade":
		// For upgrading a chart the user should be able to create, update and delete resources
		for _, v := range []string{"create", "update", "delete"} {
			actions, err := u.isAllowed(v, resources)
			if err != nil {
				return []Action{}, err
			}
			forbiddenActions = append(forbiddenActions, actions...)
		}
		if len(forbiddenActions) > 0 {
			forbiddenActions = reduceActionsByVerb(forbiddenActions)
		}
	default:
		forbiddenActions, err = u.isAllowed(action, resources)
		if err != nil {
			return []Action{}, err
		}
	}
	return forbiddenActions, nil
}

// ParseForbiddenActions parses a forbidden error returned by the Kubernetes API and return the list of forbidden actions
func ParseForbiddenActions(message string) []Action {
	// TODO(andresmgot): Helm may not return all the required permissions in the same message. At the moment of writing this
	// the only supported format is an error string so we can only parse the message with a regex
	// More info: https://github.com/helm/helm/issues/7453
	re := regexp.MustCompile(`User "(.*?)" cannot (.*?) resource "(.*?)" in API group "(.*?)"(?: in the namespace "(.*?)")?`)
	match := re.FindAllStringSubmatch(message, -1)
	forbiddenActions := []Action{}
	for _, role := range match {
		forbiddenActions = append(forbiddenActions, Action{
			// TODO(andresmgot): Return the user/serviceaccount trying to perform the action
			Verbs:       []string{role[2]},
			Resource:    role[3],
			APIVersion:  role[4],
			Namespace:   role[5],
			ClusterWide: role[5] == "",
		})
	}
	return reduceActionsByVerb(forbiddenActions)
}
