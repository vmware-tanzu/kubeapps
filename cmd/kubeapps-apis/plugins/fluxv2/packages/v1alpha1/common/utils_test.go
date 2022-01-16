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
package common

import (
	"sync"
	"testing"
)

func TestRWMutexUtils(t *testing.T) {
	rw := &sync.RWMutex{}

	writeLocked := RWMutexWriteLocked(rw)
	readLocked := RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.RLock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || !readLocked {
		t.Fatalf("expected write/read lock: [false, true], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.RUnlock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.Lock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if !writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [true, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.Unlock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}
}
