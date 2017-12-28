FROM quay.io/deis/go-dev:v1.6.0 as builder
COPY . /go/src/github.com/kubeapps/chartsvc
WORKDIR /go/src/github.com/kubeapps/chartsvc
RUN dep ensure
RUN CGO_ENABLED=0 go build -a -installsuffix cgo

FROM scratch
COPY --from=builder /go/src/github.com/kubeapps/chartsvc/chartsvc /chartsvc
EXPOSE 8080
CMD ["/chartsvc"]
