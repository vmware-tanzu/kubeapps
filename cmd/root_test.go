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

package cmd

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseObjectsSuccess(t *testing.T) {
	m1 := `apiVersion: v1
kind: Namespace
metadata:
  annotations: {}
  labels:
    name: kubeless
  name: kubeless`
	rs, err := parseObjects(m1)
	if err != nil {
		t.Error(err)
	}
	if len(rs) != 1 {
		t.Errorf("Expected 1 yaml element, got %v", len(rs))
	}

	// validate some fields of the parsed object
	if rs[0].GetAPIVersion() != "v1" {
		t.Errorf("Expected apiversion=v1, go %s", rs[0].GetAPIVersion())
	}
	if rs[0].GetKind() != "Namespace" {
		t.Errorf("Expected kind = Namespace, go %s", rs[0].GetKind())
	}
}

func TestParseObjectFailure(t *testing.T) {
	m2 := `apiVersion: v1
kind: Namespace
metadata:
    annotations: {}
  labels:
    name: kubeless
  name: kubeless`
	_, err := parseObjects(m2)
	if err == nil {
		t.Error("Expected parse fail, got success")
	}
}

func TestPopulateSecretWithPasswords(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ns := "test"
	id := "mysecret"
	err := populateSecretWithPasswords(clientset, ns, id, []string{"pass1", "pass2"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	secret, err := clientset.CoreV1().Secrets(ns).Get(id, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if secret.Data["pass1"] == nil || len(secret.Data["pass1"]) != 10 {
		t.Error("Failed to generate password")
	}
	if secret.Data["pass2"] == nil || len(secret.Data["pass2"]) != 10 {
		t.Error("Failed to generate password")
	}

	// It should noop if the secret already exists
	prevValue := secret.Data["pass1"]
	err = populateSecretWithPasswords(clientset, ns, id, []string{"pass1"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	secret2, err := clientset.CoreV1().Secrets(ns).Get(id, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if string(secret2.Data["pass1"]) != string(prevValue) {
		t.Error("It should not overwrite the value")
	}

	// If the secret already exists it should throw an error
	err = populateSecretWithPasswords(clientset, ns, id, []string{"pass3"})
	if err == nil {
		t.Error("Should throw an error if there is a conflict in the secret")
	}
}
