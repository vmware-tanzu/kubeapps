FROM quay.io/deis/go-dev:v1.6.0 as builder
COPY . /go/src/github.com/kubeapps/apprepository-controller
WORKDIR /go/src/github.com/kubeapps/apprepository-controller
RUN CGO_ENABLED=0 go build -a -installsuffix cgo

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kubeapps/apprepository-controller/apprepository-controller /apprepository-controller
CMD ["/apprepository-controller"]
