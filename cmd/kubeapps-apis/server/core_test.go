/*
Copyright © 2021 VMware
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
package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	core "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/v1"
)

func TestPluginsAvailable(t *testing.T) {
	testCases := []struct {
		name            string
		expectedPlugins []string
	}{
		{
			name:            "it returns a stubbed response currently",
			expectedPlugins: []string{"foobar.package.v1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := coreServer{}

			resp, err := cs.PluginsAvailable(context.TODO(), &core.PluginsAvailableRequest{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := resp.Plugins, tc.expectedPlugins; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
