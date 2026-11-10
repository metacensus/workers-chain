# MetaCensus Temporal server with compile-time Authorizer / ClaimMapper.
#
# TEMPORAL_VERSION must equal the go.temporal.io/server version in go.mod: the
# binary we build is linked against that library, and it has to run on the
# matching upstream image (config schema, admintools, SQL schema). CI resolves
# it from go.mod and passes it in; the default keeps local builds working.
ARG TEMPORAL_VERSION=1.30.4

FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /temporal-server ./cmd/server

FROM temporalio/server:${TEMPORAL_VERSION}
# Replace the stock binary; keep upstream ENTRYPOINT/CMD (expects `start`, TEMPORAL_* env).
COPY --from=build /temporal-server /usr/local/bin/temporal-server
