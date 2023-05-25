# syntax = docker/dockerfile:1.2

FROM golang:1.20.3 as backend

RUN apt-get install -y gcc

WORKDIR /backend-build
COPY * .
RUN --mount=type=cache,target=/opt/go \
    GOMODCACHE=/opt/go/modcache GOCACHE=/opt/go/cache \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build \
    --tags "release" \
    -ldflags "-w -s" \
    -o bytebase-unauth \
    main.go



FROM debian:bullseye-slim as monolithic

COPY --from=backend /backend-build/bytebase-unauth /usr/local/bin/
COPY --from=backend /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT [ "/usr/local/bin/bytebase-unauth" ]