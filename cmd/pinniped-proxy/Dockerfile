# Copyright 2020-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# syntax = docker/dockerfile:1

FROM rust:1.63.0 as builder

WORKDIR /pinniped-proxy
ARG VERSION

# Create a release build of pinniped-proxy itself.
COPY ./cmd/pinniped-proxy ./
ENV PINNIPED_PROXY_VERSION=$VERSION
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/pinniped-proxy/target \
    cargo build --release
RUN --mount=type=cache,target=/pinniped-proxy/target \
    cp /pinniped-proxy/target/release/pinniped-proxy /pinniped-proxy/pinniped-proxy

FROM bitnami/minideb:bullseye
RUN apt-get update && apt-get install -y ca-certificates libssl1.1 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /pinniped-proxy/pinniped-proxy /pinniped-proxy
ENV PATH="/:$PATH"

EXPOSE 3333
USER 1001
ENTRYPOINT [ "pinniped-proxy" ]
CMD [ "--help" ]
