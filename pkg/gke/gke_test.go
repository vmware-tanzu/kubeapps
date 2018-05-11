/*
Copyright (c) 2017 Bitnami

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

package gke

import (
	"os"
	"testing"
)

const (
	SuccessPath = "testsdkconfig/success"
	FailurePath = "testsdkconfig/failure"
)

func TestGetActiveUserSuccess(t *testing.T) {
	os.Setenv("CLOUDSDK_CONFIG", SuccessPath)
	email, err := GetActiveUser()
	if err != nil {
		t.Error()
	}
	if email != "foo@example.com" {
		t.Errorf("expected email = foo@example.com, got email = '%s'", email)
	}
	os.Unsetenv("CLOUDSDK_CONFIG")
}

func TestGetActiveUserFailure(t *testing.T) {
	os.Setenv("CLOUDSDK_CONFIG", FailurePath)
	email, err := GetActiveUser()
	if err == nil {
		t.Error()
	}
	if email != "" {
		t.Errorf("expected email not found, got email = %s", email)
	}
	os.Unsetenv("CLOUDSDK_CONFIG")
}

func TestBuildCrbObject(t *testing.T) {
	user := "foo"
	crd, err := BuildCrbObject(user)
	if err != nil {
		t.Errorf("expected crb object can be built successfully, got error %v", err)
	}
	if len(crd) != 1 {
		t.Errorf("expected build a single srb, got %d elements", len(crd))
	}

	data := crd[0].UnstructuredContent()
	sub := data["subjects"].([]map[string]interface{})
	if len(sub) != 1 {
		t.Errorf("expected to have only one subject, got %v", len(sub))
	}

	if sub[0]["name"].(string) != user {
		t.Errorf("expected user %s should be loaded into crb object, got %s", user, sub[0]["name"].(string))
	}
}
