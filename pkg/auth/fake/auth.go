// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	authUtils "github.com/kubeapps/kubeapps/pkg/auth"
)

type FakeAuth struct {
	ForbiddenActions []authUtils.Action
}

func (f *FakeAuth) Validate() error {
	return nil
}

func (f *FakeAuth) ValidateForNamespace(namespace string) (bool, error) {
	return true, nil
}

func (f *FakeAuth) GetForbiddenActions(namespace, action, manifest string) ([]authUtils.Action, error) {
	return f.ForbiddenActions, nil
}
