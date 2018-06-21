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

	authorizationapi "k8s.io/api/authorization/v1"
	discovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"

	yamlUtils "github.com/kubeapps/kubeapps/pkg/yaml"
)

// UserAuth contains information to check user permissions
type UserAuth struct {
	authCli      authorizationv1.AuthorizationV1Interface
	discoveryCli discovery.DiscoveryInterface
}

type resource struct {
	APIVersion string
	Kind       string
	Namespace  string
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
	authCli := kubeClient.AuthorizationV1()
	discoveryCli := kubeClient.Discovery()

	return &UserAuth{authCli, discoveryCli}, nil
}

// Validate checks if the given token is valid
func (u *UserAuth) Validate() error {
	_, err := u.authCli.SelfSubjectRulesReviews().Create(&authorizationapi.SelfSubjectRulesReview{
		Spec: authorizationapi.SelfSubjectRulesReviewSpec{
			Namespace: "default",
		},
	})
	return err
}

func resolve(discoveryCli discovery.DiscoveryInterface, groupVersion, kind string) (string, error) {
	resourceList, err := discoveryCli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return "", nil
	}
	for _, r := range resourceList.APIResources {
		if r.Kind == kind {
			return r.Name, nil
		}
	}
	return "", fmt.Errorf("Unable to find the kind %s in the resource group %s", kind, groupVersion)
}

func (u *UserAuth) canPerform(verb, group, resource, namespace string) (bool, error) {
	res, err := u.authCli.SelfSubjectAccessReviews().Create(&authorizationapi.SelfSubjectAccessReview{
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

func (u *UserAuth) isAllowed(verb string, itemsToCheck []resource) error {
	for _, i := range itemsToCheck {
		resource, err := resolve(u.discoveryCli, i.APIVersion, i.Kind)
		if err != nil {
			return err
		}
		group := i.APIVersion
		if group == "v1" {
			// The group should be empty for the core API group
			group = ""
		}
		allowed, _ := u.canPerform(verb, group, resource, i.Namespace)
		if !allowed {
			return fmt.Errorf("Unauthorized to %s %s/%s in the %s namespace", verb, i.APIVersion, resource, i.Namespace)
		}
	}
	return nil
}

// CanI returns if the user can perform the given action with the given chart and parameters
func (u *UserAuth) CanI(namespace, action, manifest string) error {
	resources, err := u.getResourcesToCheck(namespace, manifest)
	if err != nil {
		return err
	}
	switch action {
	case "upgrade":
		// For upgrading a chart the user should be able to create, update and delete resources
		for _, v := range []string{"create", "update", "delete"} {
			err = u.isAllowed(v, resources)
			if err != nil {
				return err
			}
		}
	default:
		err := u.isAllowed(action, resources)
		if err != nil {
			return err
		}
	}
	return nil
}
