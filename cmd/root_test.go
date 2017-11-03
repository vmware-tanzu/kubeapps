package cmd

import (
	"testing"
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
