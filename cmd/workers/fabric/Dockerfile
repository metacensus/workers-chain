# Build Stage

FROM golang:1.25-alpine AS buildstage

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/fabric-worker ./cmd/workers/fabric/main.go

# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/fabric-worker /bin/fabric-worker
COPY helm /helm

ENTRYPOINT ["/bin/fabric-worker"]
