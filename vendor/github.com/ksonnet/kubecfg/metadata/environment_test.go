// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

const (
	mockSpecJSON    = "spec.json"
	mockSpecJSONURI = "localhost:8080"

	mockEnvName  = "us-west/test"
	mockEnvName2 = "us-west/prod"
	mockEnvName3 = "us-east/test"
)

var mockAPIServerURI = "http://google.com"

func mockEnvironments(t *testing.T, appName string) *manager {
	spec, err := parseClusterSpec(fmt.Sprintf("file:%s", blankSwagger), testFS)
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath(appName)
	m, err := initManager(appPath, spec, &mockAPIServerURI, testFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	envNames := []string{defaultEnvName, mockEnvName, mockEnvName2, mockEnvName3}
	for _, env := range envNames {
		envPath := appendToAbsPath(m.environmentsPath, env)

		specPath := appendToAbsPath(envPath, mockSpecJSON)
		specData, err := generateSpecData(mockSpecJSONURI)
		if err != nil {
			t.Fatalf("Expected to marshal:\n%s\n, but failed", mockSpecJSONURI)
		}
		err = afero.WriteFile(testFS, string(specPath), specData, os.ModePerm)
		if err != nil {
			t.Fatalf("Could not write file at path: %s", specPath)
		}

		testDirExists(t, string(envPath))
	}

	return m
}

func testDirExists(t *testing.T, path string) {
	exists, err := afero.DirExists(testFS, path)
	if err != nil {
		t.Fatalf("Expected directory at '%s' to exist, but failed:\n%v", path, err)
	} else if !exists {
		t.Fatalf("Expected directory at '%s' to exist, but it does not", path)
	}
}

func testDirNotExists(t *testing.T, path string) {
	exists, err := afero.DirExists(testFS, path)
	if err != nil {
		t.Fatalf("Expected directory at '%s' to be removed, but failed:\n%v", path, err)
	} else if exists {
		t.Fatalf("Expected directory at '%s' to be removed, but it exists", path)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	appName := "test-delete-envs"
	m := mockEnvironments(t, appName)

	// Test that both directory and empty parent directory is deleted.
	expectedPath := appendToAbsPath(m.environmentsPath, mockEnvName3)
	parentDir := strings.Split(mockEnvName3, "/")[0]
	expectedParentPath := appendToAbsPath(m.environmentsPath, parentDir)
	err := m.DeleteEnvironment(mockEnvName3)
	if err != nil {
		t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
	}
	testDirNotExists(t, string(expectedPath))
	testDirNotExists(t, string(expectedParentPath))

	// Test that only leaf directory is deleted if parent directory is shared
	expectedPath = appendToAbsPath(m.environmentsPath, mockEnvName2)
	parentDir = strings.Split(mockEnvName2, "/")[0]
	expectedParentPath = appendToAbsPath(m.environmentsPath, parentDir)
	err = m.DeleteEnvironment(mockEnvName2)
	if err != nil {
		t.Fatalf("Expected %s to be deleted but got err:\n  %s", mockEnvName3, err)
	}
	testDirNotExists(t, string(expectedPath))
	testDirExists(t, string(expectedParentPath))
}

func TestGetEnvironments(t *testing.T) {
	m := mockEnvironments(t, "test-get-envs")

	envs, err := m.GetEnvironments()
	if err != nil {
		t.Fatalf("Expected to successfully get environments but failed:\n  %s", err)
	}

	if len(envs) != 4 {
		t.Fatalf("Expected to get %d environments, got %d", 4, len(envs))
	}

	if envs[0].URI != mockSpecJSONURI {
		t.Fatalf("Expected env URI to be %s, got %s", mockSpecJSONURI, envs[0].URI)
	}
}

func TestSetEnvironment(t *testing.T) {
	appName := "test-set-envs"
	m := mockEnvironments(t, appName)

	setName := "new-env"
	setURI := "http://example.com"
	set := Environment{Name: setName, URI: setURI}

	// Test updating an environment that doesn't exist
	err := m.SetEnvironment("notexists", &set)
	if err == nil {
		t.Fatal("Expected error when setting an environment that does not exist")
	}

	// Test updating an environment to an environment that already exists
	err = m.SetEnvironment(mockEnvName, &Environment{Name: mockEnvName2})
	if err == nil {
		t.Fatalf("Expected error when setting \"%s\" to \"%s\", because env already exists", mockEnvName, mockEnvName2)
	}

	// Test changing the name and URI of a an existing environment.
	// Ensure new env directory is created, and old directory no longer exists.
	// Also ensure URI is set in spec.json
	err = m.SetEnvironment(mockEnvName, &set)
	if err != nil {
		t.Fatalf("Could not set \"%s\", got:\n  %s", mockEnvName, err)
	}

	envPath := appendToAbsPath(AbsPath(appName), environmentsDir)
	expectedPathExists := appendToAbsPath(envPath, set.Name)
	expectedPathNotExists := appendToAbsPath(envPath, mockEnvName)

	testDirExists(t, string(expectedPathExists))
	testDirNotExists(t, string(expectedPathNotExists))

	expectedSpecPath := appendToAbsPath(expectedPathExists, specFilename)
	specData, err := afero.ReadFile(testFS, string(expectedSpecPath))
	if err != nil {
		t.Fatalf("Failed to read spec file:\n  %s", err)
	}

	var envSpec EnvironmentSpec
	err = json.Unmarshal(specData, &envSpec)
	if err != nil {
		t.Fatalf("Failed to read spec file:\n  %s", err)
	}
	if envSpec.URI != set.URI {
		t.Fatalf("Expected set URI to be \"%s\", got:\n  %s", set.URI, envSpec.URI)
	}
}
