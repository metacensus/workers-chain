# Build Stage

FROM golang:1.25-alpine AS buildstage

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/shard-worker ./cmd/workers/shard/main.go

# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/shard-worker /bin/shard-worker
COPY helm /helm

ENTRYPOINT ["/bin/shard-worker"]
