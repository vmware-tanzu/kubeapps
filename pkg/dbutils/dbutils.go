// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package dbutils

import "time"

// AssetManager basic manager for the different db types
type AssetManager interface {
	Init() error
	Close() error
}

type Config struct {
	URL      string
	Database string
	Username string
	Password string
	Timeout  time.Duration
}
