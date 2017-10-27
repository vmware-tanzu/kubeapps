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

package kubecfg

import (
	"fmt"
	"io"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ksonnet/kubecfg/metadata"
)

type EnvAddCmd struct {
	name string
	uri  string

	spec    metadata.ClusterSpec
	manager metadata.Manager
}

func NewEnvAddCmd(name, uri, specFlag string, manager metadata.Manager) (*EnvAddCmd, error) {
	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}
	log.Debugf("Generating ksonnetLib data with spec: %s", specFlag)

	return &EnvAddCmd{name: name, uri: uri, spec: spec, manager: manager}, nil
}

func (c *EnvAddCmd) Run() error {
	return c.manager.CreateEnvironment(c.name, c.uri, c.spec)
}

// ==================================================================

type EnvRmCmd struct {
	name string

	manager metadata.Manager
}

func NewEnvRmCmd(name string, manager metadata.Manager) (*EnvRmCmd, error) {
	return &EnvRmCmd{name: name, manager: manager}, nil
}

func (c *EnvRmCmd) Run() error {
	return c.manager.DeleteEnvironment(c.name)
}

// ==================================================================

type EnvListCmd struct {
	manager metadata.Manager
}

func NewEnvListCmd(manager metadata.Manager) (*EnvListCmd, error) {
	return &EnvListCmd{manager: manager}, nil
}

func (c *EnvListCmd) Run(out io.Writer) error {
	envs, err := c.manager.GetEnvironments()
	if err != nil {
		return err
	}

	// Sort environments by ascending alphabetical name
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })

	// Format each environment information for pretty printing.
	// Each environment should be outputted like the following:
	//
	//  us-west/dev     localhost:8080
	//  us-west/staging http://example.com
	//
	// To accomplish this, need to find the longest env name for proper padding.

	maxNameLen := 0
	for _, env := range envs {
		if l := len(env.Name); l > maxNameLen {
			maxNameLen = l
		}
	}

	lines := []string{}
	for _, env := range envs {
		nameSpacing := strings.Repeat(" ", maxNameLen-len(env.Name)+1)
		lines = append(lines, env.Name+nameSpacing+env.URI+"\n")
	}

	formattedEnvsList := strings.Join(lines, "")

	_, err = fmt.Fprint(out, formattedEnvsList)
	return err
}

// ==================================================================

type EnvSetCmd struct {
	name string

	desiredName string
	desiredURI  string

	manager metadata.Manager
}

func NewEnvSetCmd(name, desiredName, desiredURI string, manager metadata.Manager) (*EnvSetCmd, error) {
	return &EnvSetCmd{name: name, desiredName: desiredName, desiredURI: desiredURI, manager: manager}, nil
}

func (c *EnvSetCmd) Run() error {
	desired := metadata.Environment{Name: c.desiredName, URI: c.desiredURI}
	return c.manager.SetEnvironment(c.name, &desired)
}
