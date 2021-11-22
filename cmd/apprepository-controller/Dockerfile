# syntax = docker/dockerfile:1

FROM bitnami/golang:1.17.3 as builder
WORKDIR /go/src/github.com/kubeapps/kubeapps
COPY go.mod go.sum ./
COPY pkg pkg
COPY cmd cmd
ARG VERSION
# With the trick below, Go's build cache is kept between builds.
# https://github.com/golang/go/issues/27719#issuecomment-514747274
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-X github.com/kubeapps/kubeapps/cmd/apprepository-controller/cmd.version=$VERSION" ./cmd/apprepository-controller

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kubeapps/kubeapps/apprepository-controller /apprepository-controller
USER 1001
CMD ["/apprepository-controller"]
