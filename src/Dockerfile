FROM golang:1.22-alpine AS build

ARG TARGETARCH
ARG MOUNTPOINT_VERSION=1.8.0

RUN <<EOF
    apk update
    apk --no-cache add \
        ca-certificates \
        build-base \
        git \
        alpine-sdk \
        curl \
        libcurl \
        automake \
        autoconf \
        libxml2-dev \
        mailcap \
        fuse-dev \
        curl-dev \
        patchelf

    MP_ARCH=`echo ${TARGETARCH} | sed s/amd64/x86_64/`
    curl -L "https://s3.amazonaws.com/mountpoint-s3-release/${MOUNTPOINT_VERSION}/$MP_ARCH/mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz" -o mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz
    curl -L "https://s3.amazonaws.com/mountpoint-s3-release/${MOUNTPOINT_VERSION}/$MP_ARCH/mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz.asc" -o mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz.asc
    curl -L "https://s3.amazonaws.com/mountpoint-s3-release/public_keys/KEYS" -o KEYS

    gpg --import KEYS
    (gpg --fingerprint mountpoint-s3@amazon.com | grep "673F E406 1506 BB46 9A0E  F857 BE39 7A52 B086 DA5A")
    gpg --verify mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz.asc

    mkdir -p /mountpoint-s3
    tar -xvzf mount-s3-${MOUNTPOINT_VERSION}-$MP_ARCH.tar.gz -C /mountpoint-s3

    # set rpath for dynamic library loading
    patchelf --set-rpath '$ORIGIN' /mountpoint-s3/bin/mount-s3

#    curl -L https://github.com/awslabs/mountpoint-s3/archive/refs/tags/mountpoint-s3-1.8.0.zip -o mount-s3.zip
#    unzip mount-s3.zip
#    cd mountpoint-s3-mountpoint-s3-1.8.0
#    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && source "$HOME/.cargo/env"
#    cargo build --release
#    mv target/release/mount-s3 /usr/local/bin/mount-s3
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
COPY --from=build /mountpoint-s3/bin/mount-s3 /usr/local/bin/mount-s3

COPY src/rsyslog.conf /etc/rsyslog.conf
COPY src/fuse.conf /etc/fuse.conf
