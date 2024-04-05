# Copyright 2022-2024 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

WORKDIR = $(shell pwd)
# HUGO_VERSION should be in sync with the one set in ../../site/netlify.toml
# see https://github.com/gohugoio/hugo/releases
HUGO_VERSION = 0.124.1

# This file provides targets that helps with the development of the kubeapps.dev site.

site-server:
	docker run --rm -it -v $(WORKDIR):/src -p 1313:1313 -w /src/site "klakegg/hugo:$(HUGO_VERSION)-busybox" server --disableFastRender

.PHONY: site-server
