# Build Stage

FROM golang:1.23-alpine3.20 AS buildstage

WORKDIR /src

COPY . .

RUN pwd
RUN ls -la

RUN go mod download

RUN go build -o /bin/founder ./cmd/founder/main.go


# Deploy Stage

FROM alpine:latest

WORKDIR /

COPY --from=buildstage /bin/founder /bin/founder

ENTRYPOINT ["/bin/founder"]