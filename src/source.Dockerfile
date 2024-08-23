FROM golang:1.22-alpine AS build

ARG TARGETARCH
ARG MOUNTPOINT_VERSION=1.8.0

RUN <<EOF
    apk update
    apk --no-cache add \
        ca-certificates \
        build-base \
        gcc \
        g++ \
        git \
        alpine-sdk \
        curl \
        libcurl \
        cmake \
        make \
        clang \
        clang-dev \
        fuse \
        fuse-dev \
        curl-dev \
        zlib \
        pkgconfig

    MP_ARCH=`echo ${TARGETARCH} | sed s/amd64/x86_64/`

    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && source "$HOME/.cargo/env"

    git clone --recurse-submodules https://github.com/awslabs/mountpoint-s3.git --branch v${MOUNTPOINT_VERSION} --depth 1
    cd mountpoint-s3
    export RUSTFLAGS="-C target-feature=-crt-static"
    cargo build --release
    mv target/release/mount-s3 /usr/local/bin/mount-s3
EOF

COPY src /src/

RUN cd /src/ && CGO_ENABLED=0 GOOS=linux go build -o /docker-mount-s3

FROM alpine:latest

LABEL org.opencontainers.image.title="aekis/docker-mount-s3"
LABEL org.opencontainers.image.description="Provides a Docker Volume Driver for Mounting Amazon S3 Buckets using Mountpoint S3"
LABEL org.opencontainers.image.authors="Axel Mendoza <axel@aekis.dev>"
LABEL org.opencontainers.image.url="https://github.com/aekis-dev/docker-mount-s3"
LABEL org.opencontainers.image.documentation="https://github.com/aekis-dev/docker-mount-s3/README.md"
LABEL org.opencontainers.image.source="https://github.com/aekis-dev/docker-mount-s3/src/Dockerfile"

RUN <<EOF
    apk update
    apk --no-cache add \
        fuse \
        libcurl \
        libxml2 \
        libgcc \
        libstdc++ \
        mailcap \
        ca-certificates \
        rsyslog \
        tini
    deluser xfs
    mkdir -p /var/lib/rsyslog
    rm -rf /var/cache/apk/*
EOF

COPY --from=build /docker-mount-s3 /
COPY --from=build /usr/local/bin/mount-s3 /usr/local/bin/mount-s3

COPY src/rsyslog.conf /etc/rsyslog.conf
COPY src/fuse.conf /etc/fuse.conf
