// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"regexp"
)

// Action represents a specific set of verbs against a resource
type Action struct {
	APIVersion  string   `json:"apiGroup"`
	Resource    string   `json:"resource"`
	Namespace   string   `json:"namespace"`
	ClusterWide bool     `json:"clusterWide"`
	Verbs       []string `json:"verbs"`
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
