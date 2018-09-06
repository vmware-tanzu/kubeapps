FROM quay.io/deis/go-dev:v1.8.2 as builder
COPY . /go/src/github.com/kubeapps/kubeapps
WORKDIR /go/src/github.com/kubeapps/kubeapps
RUN CGO_ENABLED=0 go build -a -installsuffix cgo ./cmd/chart-repo

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kubeapps/kubeapps/chart-repo /chart-repo
CMD ["/chart-repo"]
