// Copyright 2017-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package signals

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
