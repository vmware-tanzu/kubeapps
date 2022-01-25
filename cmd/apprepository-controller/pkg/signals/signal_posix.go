//go:build !windows
// +build !windows

// Copyright 2017-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package signals

import (
	"os"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
