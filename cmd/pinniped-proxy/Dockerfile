FROM rust:1.47 as builder

# Create a new project and build pinniped-proxy dependencies only, so that they
# are built as a layer on their own.
RUN USER=root cargo new --bin pinniped-proxy
WORKDIR /pinniped-proxy
COPY ./cmd/pinniped-proxy/Cargo.* ./
RUN cargo build --release && rm src/*.rs

# Create a release build of pinniped-proxy itself.
ADD ./cmd/pinniped-proxy ./
RUN rm ./target/release/deps/pinniped_proxy*
RUN cargo build --release

# Put it all together on a scratch image.
# TODO: Turns out we'll need to include libssl on the scratch image, perhaps compiling libssl or copying dyn libs listed by ldd.
# FROM scratch
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# FROM bitnami/minideb:buster-snapshot-20201121T213529Z
FROM debian:10.6-slim
RUN apt-get update && apt-get install -y ca-certificates libssl1.1 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /pinniped-proxy/target/release/pinniped-proxy /pinniped-proxy

EXPOSE 3333
CMD ["/pinniped-proxy"]
