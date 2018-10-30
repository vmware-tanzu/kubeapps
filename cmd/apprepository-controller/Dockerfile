FROM golang:1.11 as builder
COPY . /go/src/github.com/kubeapps/kubeapps
WORKDIR /go/src/github.com/kubeapps/kubeapps
RUN CGO_ENABLED=0 go build -a -installsuffix cgo ./cmd/apprepository-controller

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kubeapps/kubeapps/apprepository-controller /apprepository-controller
CMD ["/apprepository-controller"]
