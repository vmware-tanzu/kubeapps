/*
Copyright 2022 The Flux authors

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

package transport

import (
	"crypto/tls"
	"testing"
)

func Test_TransportReuse(t *testing.T) {
	t1 := NewOrIdle(nil)
	t2 := NewOrIdle(nil)

	if t1 == t2 {
		t.Errorf("same transported returned twice")
	}

	err := Release(t2)
	if err != nil {
		t.Errorf("error releasing transport t2: %v", err)
	}

	t3 := NewOrIdle(&tls.Config{
		ServerName: "testing",
	})
	if t3.TLSClientConfig == nil || t3.TLSClientConfig.ServerName != "testing" {
		t.Errorf("TLSClientConfig not properly configured")
	}

	err = Release(t3)
	if err != nil {
		t.Errorf("error releasing transport t3: %v", err)
	}
	if t3.TLSClientConfig != nil {
		t.Errorf("TLSClientConfig not cleared after release")
	}

	err = Release(nil)
	if err == nil {
		t.Errorf("should not allow release nil transport")
	} else if err.Error() != "cannot release nil transport" {
		t.Errorf("wanted error message: 'cannot release nil transport' got: %q", err.Error())
	}
}
