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
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

const (
	defaultEnvName = "default"

	schemaFilename        = "swagger.json"
	extensionsLibFilename = "k.libsonnet"
	k8sLibFilename        = "k8s.libsonnet"
	specFilename          = "spec.json"
)

// Environment represents all fields of a ksonnet environment
type Environment struct {
	Path string
	Name string
	URI  string
}

// EnvironmentSpec represents the contents in spec.json.
type EnvironmentSpec struct {
	URI string `json:"uri"`
}

func (m *manager) CreateEnvironment(name, uri string, spec ClusterSpec) error {
	extensionsLibData, k8sLibData, specData, err := m.generateKsonnetLibData(spec)
	if err != nil {
		log.Debugf("Failed to write '%s'", specFilename)
		return err
	}

	err = m.createEnvironment(name, uri, extensionsLibData, k8sLibData, specData)
	if err == nil {
		log.Infof("Environment '%s' pointing to cluster at URI '%s' successfully created", name, uri)
	}
	return err
}

func (m *manager) createEnvironment(name, uri string, extensionsLibData, k8sLibData, specData []byte) error {
	exists, err := m.environmentExists(name)
	if err != nil {
		log.Debug("Failed to check whether environment exists")
		return err
	}
	if exists {
		return fmt.Errorf("Environment '%s' already exists", name)
	}

	// ensure environment name does not contain punctuation
	if !isValidName(name) {
		return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	log.Infof("Creating environment '%s' with uri '%s'", name, uri)

	envPath := appendToAbsPath(m.environmentsPath, name)
	err = m.appFS.MkdirAll(string(envPath), defaultFolderPermissions)
	if err != nil {
		return err
	}

	log.Infof("Generating environment metadata at path '%s'", envPath)

	// Generate the schema file.
	log.Debugf("Generating '%s', length: %d", schemaFilename, len(specData))
	schemaPath := appendToAbsPath(envPath, schemaFilename)
	err = afero.WriteFile(m.appFS, string(schemaPath), specData, defaultFilePermissions)
	if err != nil {
		log.Debugf("Failed to write '%s'", schemaFilename)
		return err
	}

	log.Debugf("Generating '%s', length: %d", k8sLibFilename, len(k8sLibData))
	k8sLibPath := appendToAbsPath(envPath, k8sLibFilename)
	err = afero.WriteFile(m.appFS, string(k8sLibPath), k8sLibData, defaultFilePermissions)
	if err != nil {
		log.Debugf("Failed to write '%s'", k8sLibFilename)
		return err
	}

	log.Debugf("Generating '%s', length: %d", extensionsLibFilename, len(extensionsLibData))
	extensionsLibPath := appendToAbsPath(envPath, extensionsLibFilename)
	err = afero.WriteFile(m.appFS, string(extensionsLibPath), extensionsLibData, defaultFilePermissions)
	if err != nil {
		log.Debugf("Failed to write '%s'", extensionsLibFilename)
		return err
	}

	// Generate the environment spec file.
	envSpecData, err := generateSpecData(uri)
	if err != nil {
		return err
	}

	log.Debugf("Generating '%s', length: %d", specFilename, len(envSpecData))
	envSpecPath := appendToAbsPath(envPath, specFilename)
	return afero.WriteFile(m.appFS, string(envSpecPath), envSpecData, defaultFilePermissions)
}

func (m *manager) DeleteEnvironment(name string) error {
	envPath := string(appendToAbsPath(m.environmentsPath, name))

	// Check whether this environment exists
	envExists, err := m.environmentExists(name)
	if err != nil {
		log.Debug("Failed to check whether environment exists")
		return err
	}
	if !envExists {
		return fmt.Errorf("Environment '%s' does not exist", name)
	}

	log.Infof("Deleting environment '%s' at path '%s'", name, envPath)

	// Remove the directory and all files within the environment path.
	err = m.appFS.RemoveAll(envPath)
	if err != nil {
		log.Debugf("Failed to remove environment directory at path '%s'", envPath)
		return err
	}

	// Need to ensure empty parent directories are also removed.
	log.Debug("Removing empty parent directories, if any")
	parentDir := name
	for parentDir != "." {
		parentDir = filepath.Dir(parentDir)
		parentPath := string(appendToAbsPath(m.environmentsPath, parentDir))

		isEmpty, err := afero.IsEmpty(m.appFS, parentPath)
		if err != nil {
			log.Debugf("Failed to check whether parent directory at path '%s' is empty", parentPath)
			return err
		}
		if isEmpty {
			log.Debugf("Failed to remove parent directory at path '%s'", parentPath)
			err := m.appFS.RemoveAll(parentPath)
			if err != nil {
				return err
			}
		}
	}

	log.Infof("Successfully removed environment '%s'", name)
	return nil
}

func (m *manager) GetEnvironments() ([]*Environment, error) {
	envs := []*Environment{}

	log.Info("Retrieving all environments")
	err := afero.Walk(m.appFS, string(m.environmentsPath), func(path string, f os.FileInfo, err error) error {
		isDir, err := afero.IsDir(m.appFS, path)
		if err != nil {
			log.Debugf("Failed to check whether the path at '%s' is a directory", path)
			return err
		}

		if isDir {
			// Only want leaf directories containing a spec.json
			specPath := filepath.Join(path, specFilename)
			specFileExists, err := afero.Exists(m.appFS, specPath)
			if err != nil {
				log.Debugf("Failed to check whether spec file at '$s' exists", specPath)
				return err
			}
			if specFileExists {
				envName := filepath.Clean(strings.TrimPrefix(path, string(m.environmentsPath)+"/"))
				specFile, err := afero.ReadFile(m.appFS, specPath)
				if err != nil {
					log.Debugf("Failed to read spec file at path '%s'", specPath)
					return err
				}
				var envSpec EnvironmentSpec
				err = json.Unmarshal(specFile, &envSpec)
				if err != nil {
					log.Debugf("Failed to convert the spec file at path '%s' to JSON", specPath)
					return err
				}

				log.Debugf("Found environment '%s', with uri '%s", envName, envSpec.URI)
				envs = append(envs, &Environment{Name: envName, Path: path, URI: envSpec.URI})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return envs, nil
}

func (m *manager) GetEnvironment(name string) (*Environment, error) {
	envs, err := m.GetEnvironments()
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		if env.Name == name {
			return env, nil
		}
	}

	return nil, fmt.Errorf("Environment '%s' does not exist", name)
}

func (m *manager) SetEnvironment(name string, desired *Environment) error {
	// Check whether this environment exists
	envExists, err := m.environmentExists(name)
	if err != nil {
		log.Debugf("Failed to check whether '%s' exists", name)
		return err
	}
	if !envExists {
		return fmt.Errorf("Environment '%s' does not exist", name)
	}

	// If the name has changed, the directory location needs to be moved to
	// reflect the change.
	if name != desired.Name && len(desired.Name) != 0 {
		// ensure new environment name does not contain punctuation
		if !isValidName(desired.Name) {
			return fmt.Errorf("Environment name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
		}

		log.Infof("Setting environment name from '%s' to '%s'", name, desired.Name)

		// Ensure not overwriting another environment
		desiredExists, err := m.environmentExists(desired.Name)
		if err != nil {
			log.Debugf("Failed to check whether environment '%s' already exists", desired.Name)
			return err
		}
		if desiredExists {
			return fmt.Errorf("Can not update '%s' to '%s', it already exists", name, desired.Name)
		}

		// Move the directory
		pathOld := string(appendToAbsPath(m.environmentsPath, name))
		pathNew := string(appendToAbsPath(m.environmentsPath, desired.Name))
		log.Debugf("Moving directory at path '%s' to '%s'", pathOld, pathNew)
		err = m.appFS.Rename(pathOld, pathNew)
		if err != nil {
			log.Debugf("Failed to move path '%s' to '%s", pathOld, pathNew)
			return err
		}

		name = desired.Name
	}

	// Update fields in spec.json
	if len(desired.URI) != 0 {
		log.Infof("Setting environment URI to '%s'", desired.URI)

		newSpec, err := generateSpecData(desired.URI)
		if err != nil {
			log.Debugf("Failed to generate %s with URI '%s'", specFilename, desired.URI)
			return err
		}

		envPath := appendToAbsPath(m.environmentsPath, name)
		specPath := appendToAbsPath(envPath, specFilename)

		err = afero.WriteFile(m.appFS, string(specPath), newSpec, defaultFilePermissions)
		if err != nil {
			log.Debugf("Failed to write %s at path '%s'", specFilename, specPath)
			return err
		}
	}

	log.Infof("Successfully updated environment '%s'", name)
	return nil
}

func (m *manager) generateKsonnetLibData(spec ClusterSpec) ([]byte, []byte, []byte, error) {
	// Get cluster specification data, possibly from the network.
	text, err := spec.data()
	if err != nil {
		return nil, nil, nil, err
	}

	ksonnetLibDir := appendToAbsPath(m.environmentsPath, defaultEnvName)

	// Deserialize the API object.
	s := kubespec.APISpec{}
	err = json.Unmarshal(text, &s)
	if err != nil {
		return nil, nil, nil, err
	}

	s.Text = text
	s.FilePath = filepath.Dir(string(ksonnetLibDir))

	// Emit Jsonnet code.
	extensionsLibData, k8sLibData, err := ksonnet.Emit(&s, nil, nil)
	return extensionsLibData, k8sLibData, text, err
}

func generateSpecData(uri string) ([]byte, error) {
	// Format the spec json and return; preface keys with 2 space idents.
	return json.MarshalIndent(EnvironmentSpec{URI: uri}, "", "  ")
}

func (m *manager) environmentExists(name string) (bool, error) {
	envs, err := m.GetEnvironments()
	if err != nil {
		return false, err
	}

	envExists := false
	for _, env := range envs {
		if env.Name == name {
			envExists = true
			break
		}
	}

	return envExists, nil
}
