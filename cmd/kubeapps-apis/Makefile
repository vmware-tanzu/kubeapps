# Copyright 2021 the Kubeapps contributors. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Ensure the required version of cli tooling in tools/tools.go is installed.
# The versions are stored in ../go.mod as usual and this same approach is used in
# both octant and pinniped for build tooling.
# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
cli-dependencies:
	go get \
		github.com/spf13/cobra/cobra

run:
	go run main.go

.PHONY: cli-dependencies run
