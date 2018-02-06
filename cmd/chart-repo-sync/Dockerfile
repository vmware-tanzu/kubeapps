FROM quay.io/deis/go-dev:v1.6.0 as builder
COPY . /go/src/github.com/kubernetes-helm/monocular/src/api/cmd/chart-repo-sync
WORKDIR /go/src/github.com/kubernetes-helm/monocular/src/api/cmd/chart-repo-sync
RUN dep ensure
RUN CGO_ENABLED=0 go build -a -installsuffix cgo

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kubernetes-helm/monocular/src/api/cmd/chart-repo-sync/chart-repo-sync /chart-repo-sync
CMD ["/chart-repo-sync"]
