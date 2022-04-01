# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

FROM bitnami/node:16.14.2 AS build
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn install --frozen-lockfile

RUN mkdir /app/src
COPY . /app
RUN yarn run prettier-check && yarn run ts-compile-check
RUN yarn run build

FROM bitnami/nginx:1.21-debian-10
COPY --from=build /app/build /app
