FROM alpine:3.6 as builder
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./tiller-proxy-static /proxy
CMD ["/proxy"]
