# syntax=docker/dockerfile:1

# Build natively and cross-compile: no QEMU per target architecture.
# CGO_ENABLED=0 is what golang:alpine already did (it ships no C compiler).
FROM --platform=$BUILDPLATFORM golang:1.25.8-alpine AS buildstage
WORKDIR /src
ARG TARGETOS TARGETARCH
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /bin/shard-worker ./cmd/workers/shard

FROM alpine:3.23
COPY --from=buildstage /bin/shard-worker /bin/shard-worker
COPY helm /helm
ENTRYPOINT ["/bin/shard-worker"]
