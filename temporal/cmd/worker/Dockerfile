# Build Stage

FROM golang:1.23-alpine3.20 AS buildstage

WORKDIR /src

COPY . .

RUN pwd
RUN ls -la

RUN go mod download

RUN go build -o /bin/worker ./cmd/worker/main.go


# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/worker /bin/worker

ENTRYPOINT ["/bin/worker"]