# Build Stage

FROM golang:1.25-alpine AS buildstage

WORKDIR /src/temporal

# Copy go.mod and go.sum first for better layer caching
# Dependencies are only re-downloaded when these files change
COPY temporal/go.mod temporal/go.sum ./

RUN go mod download

# Now copy the rest of the source code
COPY temporal/ .

RUN go build -o /bin/shard-worker ./cmd/workers/shard/main.go


# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/shard-worker /bin/shard-worker
# Copy Helm charts for deployment activities
COPY helm /helm

ENTRYPOINT ["/bin/shard-worker"]
