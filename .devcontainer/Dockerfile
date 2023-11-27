FROM cosmwasm/wasmd:v0.45.0 as wasmd

FROM cosmwasm/rust-optimizer:0.45 as rust-optimizer

FROM mcr.microsoft.com/devcontainers/rust:1-bullseye

COPY --from=wasmd /usr/bin/wasmd /usr/local/bin/wasmd
COPY --from=wasmd /opt/* /opt/

COPY pond /usr/local/bin/

# Install additional packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends curl jq \
    && apt-get -y install ca-certificates curl python3-pip python3-dev python3-setuptools python3-wheel \
    && sudo rm -rf /var/lib/apt/lists/*

RUN rustup update stable \
    && rustup target add wasm32-unknown-unknown

RUN cargo install cargo-run-script cargo-tarpaulin

# RUN pond init --nodes 1