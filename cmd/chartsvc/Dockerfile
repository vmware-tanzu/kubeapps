FROM quay.io/deis/go-dev:v1.8.2 as builder
COPY . /go/src/github.com/kubeapps/kubeapps
WORKDIR /go/src/github.com/kubeapps/kubeapps
RUN CGO_ENABLED=0 go build -a -installsuffix cgo ./cmd/chartsvc

FROM scratch
COPY --from=builder /go/src/github.com/kubeapps/kubeapps/chartsvc /chartsvc
EXPOSE 8080
CMD ["/chartsvc"]
