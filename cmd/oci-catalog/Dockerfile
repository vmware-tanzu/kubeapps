# Copyright 2023-2024 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# syntax = docker/dockerfile:1

FROM rust:1.81.0 AS builder

WORKDIR /oci-catalog
ARG VERSION="devel"
ARG TARGETARCH

# https://github.com/grpc-ecosystem/grpc-health-probe/releases/
ARG GRPC_HEALTH_PROBE_VERSION="0.4.34"

# Ensure protoc is available for the build.rs step.
RUN apt-get update && apt-get -y install --no-install-recommends protobuf-compiler && rm -rf /var/lib/apt/lists/*

# TODO: Remove and instead use built-in gRPC container probes once we're supporting >= 1.24 only. https://kubernetes.io/blog/2022/05/13/grpc-probes-now-in-beta/
RUN curl -sSL "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-${TARGETARCH}" -o "/bin/grpc_health_probe" && chmod +x "/bin/grpc_health_probe"

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

FROM bitnami/minideb:bookworm
RUN apt-get update && apt-get install -y ca-certificates libssl3 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /oci-catalog/oci-catalog /oci-catalog
COPY --from=builder /bin/grpc_health_probe /bin/
USER 1001
EXPOSE 50061
CMD [ "/oci-catalog"]
