// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/vmware-tanzu/kubeapps/pkg/helm"
)

// CheckHeader verifies that the given puller contains the given header
func CheckHeader(t *testing.T, puller helm.ChartPuller, key, value string) {
	// The header property is private so we need to use reflect to get its value
	resolver := puller.(*helm.OCIPuller).Resolver
	resolverValue := reflect.ValueOf(resolver)
	headerValue := reflect.Indirect(resolverValue).FieldByName("header")
	got := fmt.Sprintf("%v", headerValue)
	expected := fmt.Sprintf("%s:[%s]", key, value)
	if !strings.Contains(got, expected) {
		t.Errorf("Expecting %s to contain %s", got, expected)
	}
}
