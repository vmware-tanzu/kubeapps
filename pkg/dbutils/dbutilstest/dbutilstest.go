// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package dbutilstest

import (
	"os"
	"strconv"
	"testing"
)

const KubeappsTestNamespace = "kubeapps"

func IsEnvVarTrue(t *testing.T, envvar string) bool {
	enableEnvVar := os.Getenv(envvar)
	isTrue := false
	if enableEnvVar != "" {
		var err error
		isTrue, err = strconv.ParseBool(enableEnvVar)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
	return isTrue
}
