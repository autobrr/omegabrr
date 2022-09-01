# build app
FROM golang:1.19-alpine3.16 AS app-builder

ARG VERSION=dev
ARG REVISION=dev
ARG BUILDTIME

RUN apk add --no-cache git make build-base

ENV SERVICE=example

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

#ENV GOOS=linux
#ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${REVISION} -X main.date=${BUILDTIME}" -o bin/example cmd/example/main.go

# build runner
FROM alpine:latest

LABEL org.opencontainers.image.source = "https://github.com/autobrr/example"

RUN useradd -r -u 1001 -g appuser appuser
USER appuser

ENV HOME="/config" \
XDG_CONFIG_HOME="/config" \
XDG_DATA_HOME="/config"

RUN apk --no-cache add ca-certificates curl

COPY --from=app-builder /src/bin/example /usr/local/bin/

WORKDIR /config

VOLUME /config

EXPOSE 4400

ENTRYPOINT ["/usr/local/bin/example", "run", "--config", "/config"]
#CMD ["--config", "/config"]
