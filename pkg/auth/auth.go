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
	"fmt"
	"strings"

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
	Validate() error
	GetResourceList(groupVersion string) (*metav1.APIResourceList, error)
	CanI(verb, group, resource, namespace string) (bool, error)
}

type k8sAuth struct {
	AuthCli      authorizationv1.AuthorizationV1Interface
	DiscoveryCli discovery.DiscoveryInterface
}

func (u k8sAuth) Validate() error {
	_, err := u.AuthCli.SelfSubjectRulesReviews().Create(&authorizationapi.SelfSubjectRulesReview{
		Spec: authorizationapi.SelfSubjectRulesReviewSpec{
			Namespace: "default",
		},
	})
	return err
}

func (u k8sAuth) GetResourceList(groupVersion string) (*metav1.APIResourceList, error) {
	return u.DiscoveryCli.ServerResourcesForGroupVersion(groupVersion)
}

func (u k8sAuth) CanI(verb, group, resource, namespace string) (bool, error) {
	res, err := u.AuthCli.SelfSubjectAccessReviews().Create(&authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Group:     group,
				Resource:  resource,
				Verb:      verb,
				Namespace: namespace,
			},
		},
	})
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
	Validate() error
	GetForbiddenActions(namespace, action, manifest string) ([]Action, error)
}

// NewAuth creates an auth agent
func NewAuth(token string) (*UserAuth, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Overwrite default token
	config.BearerToken = token
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

// Validate checks if the given token is valid
func (u *UserAuth) Validate() error {
	return u.k8sAuth.Validate()
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

func reduceActionsByVerb(actions []Action) []Action {
	resMap := map[string]Action{}
	res := []Action{}
	for _, action := range actions {
		req := fmt.Sprintf("%s/%s/%s", action.Namespace, action.APIVersion, action.Resource)
		if _, ok := resMap[req]; ok {
			// Element already exists
			verbs := append(resMap[req].Verbs, action.Verbs...)
			resMap[req] = Action{
				APIVersion: action.APIVersion,
				Resource:   action.Resource,
				Namespace:  action.Namespace,
				Verbs:      verbs,
			}
		} else {
			resMap[req] = action
		}
	}
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
