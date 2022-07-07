# Copyright 2020-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# syntax = docker/dockerfile:1

FROM bitnami/golang:1.18.3 as builder
WORKDIR /go/src/github.com/vmware-tanzu/kubeapps
COPY go.mod go.sum ./
COPY pkg pkg
COPY cmd cmd
ARG VERSION

ARG GOSEC_VERSION="2.12.0"
RUN curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v$GOSEC_VERSION

# Run gosec to detect any security-related error at build time
RUN gosec ./cmd/asset-syncer/...
RUN gosec ./pkg/...

# With the trick below, Go's build cache is kept between builds.
# https://github.com/golang/go/issues/27719#issuecomment-514747274
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-X github.com/vmware-tanzu/kubeapps/cmd/asset-syncer/cmd.version=$VERSION" ./cmd/asset-syncer

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/vmware-tanzu/kubeapps/asset-syncer /asset-syncer
USER 1001
CMD ["/asset-syncer"]
