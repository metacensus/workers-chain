# Build Stage

FROM golang:1.25-alpine AS buildstage

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/infra-worker ./cmd/workers/infra/main.go

# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/infra-worker /bin/infra-worker
COPY helm /helm

ENTRYPOINT ["/bin/infra-worker"]
