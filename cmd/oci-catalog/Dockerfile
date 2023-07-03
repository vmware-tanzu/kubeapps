# Copyright 2023 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# syntax = docker/dockerfile:1

FROM rust:1.70.0 as builder

WORKDIR /oci-catalog
ARG VERSION

# Ensure protoc is available for the build.rs step.
RUN apt-get update && apt-get -y install --no-install-recommends protobuf-compiler && rm -rf /var/lib/apt/lists/*

# Create a release build of oci-catalog itself.
# To build a statically linked version, will need to use the rusttls rather than
# OS-native,
# reqwest = { version ="0.11", default_features=false, features = ["rustls-tls"] }
# then build for the linux-musl conditionally with something like
# https://stackoverflow.com/a/75350455 in the Dockerfile to set a target
# linux-musl target specific for the arch.
COPY ./cmd/oci-catalog ./
ENV OCI_CATALOG_VERSION=$VERSION
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/oci-catalog/target \
    cargo build --release
RUN --mount=type=cache,target=/oci-catalog/target \
    cp /oci-catalog/target/release/oci-catalog /oci-catalog/oci-catalog

FROM bitnami/minideb:bullseye
RUN apt-get update && apt-get install -y ca-certificates libssl1.1 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /oci-catalog/oci-catalog /oci-catalog
USER 1001
EXPOSE 50051
CMD [ "/oci-catalog"]
