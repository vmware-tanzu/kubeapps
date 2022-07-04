# Copyright 2019-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

FROM mcr.microsoft.com/playwright:v1.23.0-focal
WORKDIR  /app/

# Copy and install deps
COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

# Install browsers
RUN npx playwright install

# Note that the playwright config and the actual test files
# will be later passed via kubectl cp in runtime
