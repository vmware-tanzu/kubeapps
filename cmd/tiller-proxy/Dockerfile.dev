FROM alpine:3.6
RUN apk --no-cache add ca-certificates
COPY ./proxy-static /proxy
CMD ["/proxy"]
