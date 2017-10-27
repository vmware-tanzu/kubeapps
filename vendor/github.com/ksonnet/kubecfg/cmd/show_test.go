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

package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func cmdOutput(t *testing.T, args []string) string {
	var buf bytes.Buffer
	RootCmd.SetOutput(&buf)
	defer RootCmd.SetOutput(nil)

	t.Log("Running args", args)
	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		t.Fatal("command failed:", err)
	}

	return buf.String()
}

func TestShow(t *testing.T) {
	formats := map[string]func(string) (interface{}, error){
		"json": func(text string) (ret interface{}, err error) {
			err = json.Unmarshal([]byte(text), &ret)
			return
		},

		/* Temporarily(!) disabled due to
		   https://github.com/ksonnet/kubecfg/issues/99
		"yaml": func(text string) (ret interface{}, err error) {
			err = yaml.Unmarshal([]byte(text), &ret)
			return
		},
		*/
	}

	// Use the fact that JSON is also valid YAML ..
	expected := `
{
  "apiVersion": "v0alpha1",
  "kind": "TestObject",
  "nil": null,
  "bool": true,
  "number": 42,
  "string": "bar",
  "notAVal": "aVal",
  "notAnotherVal": "aVal2",
  "filevar": "foo\n",
  "array": ["one", 2, [3]],
  "object": {"foo": "bar"}
}
`

	for format, parser := range formats {
		expected, err := parser(expected)
		if err != nil {
			t.Errorf("error parsing *expected* value: %s", err)
		}

		os.Setenv("anVar", "aVal2")
		defer os.Unsetenv("anVar")

		output := cmdOutput(t, []string{"show",
			"-J", filepath.FromSlash("../testdata/lib"),
			"-o", format,
			"-f", filepath.FromSlash("../testdata/test.jsonnet"),
			"-V", "aVar=aVal",
			"-V", "anVar",
			"--ext-str-file", "filevar=" + filepath.FromSlash("../testdata/extvar.file"),
		})

		t.Log("output is", output)
		actual, err := parser(output)
		if err != nil {
			t.Errorf("error parsing output of format %s: %s", format, err)
		} else if !reflect.DeepEqual(expected, actual) {
			t.Errorf("format %s expected != actual: %s != %s", format, expected, actual)
		}
	}
}
