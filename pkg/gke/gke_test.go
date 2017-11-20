package gke

import (
	"testing"
)

const (
	SUCCESS_PATH = "testsdkconfig/success"
	FAILURE_PATH = "testsdkconfig/failure"
)

func TestGetActiveUserSuccess(t *testing.T) {
	email, err := GetActiveUser(SUCCESS_PATH)
	if err != nil {
		t.Error()
	}
	if email != "foo@example.com" {
		t.Errorf("expected email = foo@example.com, got email = %s", email)
	}
}

func TestGetActiveUserFailure(t *testing.T) {
	email, err := GetActiveUser(FAILURE_PATH)
	if err == nil {
		t.Error()
	}
	if email != "" {
		t.Errorf("expected email not found, got email = %s", email)
	}
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
