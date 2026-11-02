# syntax=docker/dockerfile:1

# Build natively and cross-compile: no QEMU per target architecture.
# CGO_ENABLED=0 is what golang:alpine already did (it ships no C compiler).
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine AS buildstage
WORKDIR /src
ARG TARGETOS TARGETARCH VERSION
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-X github.com/metacensus/workers-chain/lib/workers/config.Version=${VERSION}" -o /bin/infra-worker ./cmd/workers/infra

FROM alpine:3.23
COPY --from=buildstage /bin/infra-worker /bin/infra-worker
COPY helm /helm
ENTRYPOINT ["/bin/infra-worker"]
