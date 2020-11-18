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
RUN curl https://github.com/vmware-tanzu/pinniped/releases/download/v0.2.0/pinniped-cli-linux-amd64 -o /pinniped
RUN chmod +x /pinniped

# Put it all together on a scratch image.
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pinniped-proxy/target/release/pinniped-proxy /pinniped-proxy
COPY --from=builder /pinniped /pinniped

EXPOSE 3333
CMD ["/pinniped-proxy"]
