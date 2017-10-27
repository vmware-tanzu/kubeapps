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

package prototype

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		jsonnet  string
		expected string
	}{
		// Test multiple import param replacement in a Jsonnet file.
		{
			`
			// apiVersion: 0.1
			// name: simple-service
			// description: Generates a simple service with a port exposed
			
			local k = import 'ksonnet.beta.2/k.libsonnet';
			
			local service = k.core.v1.service;
			local servicePort = k.core.v1.service.mixin.spec.portsType;
			local port = servicePort.new((import 'param://port'), (import 'param://portName'));
			
			local name = import 'param://name';
			k.core.v1.service.new('%s-service' % [name], {app: name}, port)`,

			`
			// apiVersion: 0.1
			// name: simple-service
			// description: Generates a simple service with a port exposed
			
			local k = import 'ksonnet.beta.2/k.libsonnet';
			
			local service = k.core.v1.service;
			local servicePort = k.core.v1.service.mixin.spec.portsType;
			local port = servicePort.new((${port}), (${portName}));
			
			local name = ${name};
			k.core.v1.service.new('%s-service' % [name], {app: name}, port)`,
		},
		// Test where an import param is split over multiple lines.
		{
			`
			local f = (
				import 
				// foo comment
				'param://f'
			);
			{ foo: f, }`,

			`
			local f = (
				${f}


			);
			{ foo: f, }`,
		},
		// Test where no parameters are found.
		{
			`local f = f;
			{ foo: f, }`,
			`local f = f;
			{ foo: f, }`,
		},
	}

	errors := []string{
		// Expect error where param isn't provided.
		`local f = (import 'param://');
		{ foo: f, }`,
	}

	for _, s := range tests {
		parsed, err := parse("test", s.jsonnet)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if parsed != s.expected {
			t.Errorf("Wrong conversion\n  expected: %v\n  got: %v", s.expected, parsed)
		}
	}

	for _, e := range errors {
		parsed, err := parse("test", e)
		if err == nil {
			t.Errorf("Expected error but not found\n  input: %v  got: %v", e, parsed)
		}
	}
}
