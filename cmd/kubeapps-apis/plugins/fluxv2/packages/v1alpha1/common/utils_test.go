// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
