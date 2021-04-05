FROM rust:1.51 as builder

# Create a new project and build pinniped-proxy dependencies only, so that they
# are built as a layer on their own, to reduce build times when only our code
# changes.
RUN USER=root cargo new --bin pinniped-proxy
WORKDIR /pinniped-proxy
COPY ./cmd/pinniped-proxy/Cargo.* ./
RUN cargo build --release && rm src/*.rs

# Create a release build of pinniped-proxy itself.
ADD ./cmd/pinniped-proxy ./
RUN rm ./target/release/deps/pinniped_proxy*
RUN cargo build --release

FROM bitnami/minideb:buster
RUN apt-get update && apt-get install -y ca-certificates libssl1.1 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /pinniped-proxy/target/release/pinniped-proxy /pinniped-proxy
ENV PATH="/:$PATH"

EXPOSE 3333
USER 1001
ENTRYPOINT [ "pinniped-proxy" ]
CMD [ "--help" ]
