# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

WORKDIR = $(shell pwd)
HUGO_VERSION = "0.101.0"

# This file provides targets that helps with the development of the kubeapps.dev site.

site-server:
	docker run --rm -it -v $(WORKDIR):/src -p 1313:1313 -w /src/site "klakegg/hugo:$(HUGO_VERSION)" server --disableFastRender

.PHONY: site-server
