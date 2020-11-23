FROM rust:1.47 as builder

# Create a new project and build pinniped-proxy dependencies only, so that they
# are built as a layer on their own.
RUN USER=root cargo new --bin pinniped-proxy
WORKDIR /pinniped-proxy
COPY ./Cargo.* ./
RUN cargo build --release && rm src/*.rs

# Create a release build of pinniped-proxy itself.
ADD . ./
RUN rm ./target/release/deps/pinniped_proxy*
RUN cargo build --release

# Download the released pinniped CLI (currently used to exchange creds, rather than API).
RUN curl -L https://github.com/vmware-tanzu/pinniped/releases/download/v0.2.0/pinniped-cli-linux-amd64 -o /pinniped
RUN chmod +x /pinniped

# Put it all together on a scratch image.
# TODO: Turns out we'll need to include libssl on the scratch image, perhaps compiling libssl or copying dyn libs listed by ldd.
# FROM scratch
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# FROM bitnami/minideb:buster-snapshot-20201121T213529Z
FROM debian:10.6-slim
RUN apt-get update && apt-get install -y ca-certificates libssl1.1 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /pinniped-proxy/target/release/pinniped-proxy /pinniped-proxy
COPY --from=builder /pinniped /pinniped

EXPOSE 3333
CMD ["/pinniped-proxy"]
