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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ksonnet/kubecfg/prototype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func appendToAbsPath(originalPath AbsPath, toAppend ...string) AbsPath {
	paths := append([]string{string(originalPath)}, toAppend...)
	return AbsPath(path.Join(paths...))
}

const (
	ksonnetDir      = ".ksonnet"
	libDir          = "lib"
	componentsDir   = "components"
	environmentsDir = "environments"
	vendorDir       = "vendor"
)

type manager struct {
	appFS afero.Fs

	rootPath         AbsPath
	ksonnetPath      AbsPath
	libPath          AbsPath
	componentsPath   AbsPath
	environmentsPath AbsPath
	vendorDir        AbsPath
}

func findManager(abs AbsPath, appFS afero.Fs) (*manager, error) {
	var lastBase string
	currBase := string(abs)

	for {
		currPath := path.Join(currBase, ksonnetDir)
		exists, err := afero.Exists(appFS, currPath)
		if err != nil {
			return nil, err
		}
		if exists {
			return newManager(AbsPath(currBase), appFS), nil
		}

		lastBase = currBase
		currBase = filepath.Dir(currBase)
		if lastBase == currBase {
			return nil, fmt.Errorf("No ksonnet application found")
		}
	}
}

func initManager(rootPath AbsPath, spec ClusterSpec, serverURI *string, appFS afero.Fs) (*manager, error) {
	m := newManager(rootPath, appFS)

	// Generate the program text for ksonnet-lib.
	//
	// IMPLEMENTATION NOTE: We get the cluster specification and generate
	// ksonnet-lib before initializing the directory structure so that failure of
	// either (e.g., GET'ing the spec from a live cluster returns 404) does not
	// result in a partially-initialized directory structure.
	//
	extensionsLibData, k8sLibData, specData, err := m.generateKsonnetLibData(spec)
	if err != nil {
		return nil, err
	}

	// Initialize directory structure.
	if err := m.createAppDirTree(); err != nil {
		return nil, err
	}

	// Initialize environment, and cache specification data.
	// TODO the URI for the default environment needs to be generated from KUBECONFIG
	if serverURI != nil {
		err := m.createEnvironment(defaultEnvName, *serverURI, extensionsLibData, k8sLibData, specData)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func newManager(rootPath AbsPath, appFS afero.Fs) *manager {
	return &manager{
		appFS: appFS,

		rootPath:         rootPath,
		ksonnetPath:      appendToAbsPath(rootPath, ksonnetDir),
		libPath:          appendToAbsPath(rootPath, libDir),
		componentsPath:   appendToAbsPath(rootPath, componentsDir),
		environmentsPath: appendToAbsPath(rootPath, environmentsDir),
		vendorDir:        appendToAbsPath(rootPath, vendorDir),
	}
}

func (m *manager) Root() AbsPath {
	return m.rootPath
}

func (m *manager) ComponentPaths() (AbsPaths, error) {
	paths := AbsPaths{}
	err := afero.Walk(m.appFS, string(m.componentsPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (m *manager) CreateComponent(name string, text string, templateType prototype.TemplateType) error {
	if !isValidName(name) || strings.Contains(name, "/") {
		return fmt.Errorf("Component name '%s' is not valid; must not contain punctuation, spaces, or begin or end with a slash", name)
	}

	componentPath := string(appendToAbsPath(m.componentsPath, name))
	switch templateType {
	case prototype.YAML:
		componentPath = componentPath + ".yaml"
	case prototype.JSON:
		componentPath = componentPath + ".json"
	case prototype.Jsonnet:
		componentPath = componentPath + ".jsonnet"
	default:
		return fmt.Errorf("Unrecognized prototype template type '%s'", templateType)
	}

	if exists, err := afero.Exists(m.appFS, componentPath); exists {
		return fmt.Errorf("Component with name '%s' already exists", name)
	} else if err != nil {
		return fmt.Errorf("Could not check whether component '%s' exists:\n\n%v", name, err)
	}

	log.Infof("Writing component at '%s/%s'", componentsDir, name)

	return afero.WriteFile(m.appFS, componentPath, []byte(text), defaultFilePermissions)
}

func (m *manager) LibPaths(envName string) (libPath, envLibPath AbsPath) {
	return m.libPath, appendToAbsPath(m.environmentsPath, envName)
}

func (m *manager) createAppDirTree() error {
	exists, err := afero.DirExists(m.appFS, string(m.rootPath))
	if err != nil {
		return fmt.Errorf("Could not check existance of directory '%s':\n%v", m.rootPath, err)
	} else if exists {
		return fmt.Errorf("Could not create app; directory '%s' already exists", m.rootPath)
	}

	paths := []AbsPath{
		m.rootPath,
		m.ksonnetPath,
		m.libPath,
		m.componentsPath,
		m.environmentsPath,
		m.vendorDir,
	}

	for _, p := range paths {
		if err := m.appFS.MkdirAll(string(p), defaultFolderPermissions); err != nil {
			return err
		}
	}

	return nil
}
