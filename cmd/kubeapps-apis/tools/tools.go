// +build tools
// Copyright 2021 the Kubeapps contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tools

// This file is not intended to be compiled.  Because some of these imports are
// not actual go packages, we use a build constraint at the top of this file to
// prevent tools from inspecting the imports.

import (
	_ "github.com/spf13/cobra/cobra"
)
