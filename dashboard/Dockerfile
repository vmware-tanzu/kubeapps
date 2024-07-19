# Copyright 2018-2024 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

FROM bitnami/node:20.15.1 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
COPY . /app
RUN yarn run prettier-check && yarn run ts-compile-check
RUN yarn run build

FROM bitnami/nginx:1.27.0
COPY --from=build /app/build /app
