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
	"encoding/json"
	"fmt"
	"io"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ShowCmd represents the show subcommand
type ShowCmd struct {
	Format string
}

func (c ShowCmd) Run(apiObjects []*unstructured.Unstructured, out io.Writer) error {
	switch c.Format {
	case "yaml":
		for _, obj := range apiObjects {
			fmt.Fprintln(out, "---")
			// Urgh.  Go via json because we need
			// to trigger the custom scheme
			// encoding.
			buf, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			o := map[string]interface{}{}
			if err := json.Unmarshal(buf, &o); err != nil {
				return err
			}
			buf, err = yaml.Marshal(o)
			if err != nil {
				return err
			}
			out.Write(buf)
		}
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		for _, obj := range apiObjects {
			// TODO: this is not valid framing for JSON
			if len(apiObjects) > 1 {
				fmt.Fprintln(out, "---")
			}
			if err := enc.Encode(obj); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unknown --format: %s", c.Format)
	}

	return nil
}
