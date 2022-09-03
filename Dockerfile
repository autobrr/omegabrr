# build app
FROM golang:1.19-alpine3.16 AS app-builder

ARG VERSION=dev
ARG REVISION=dev
ARG BUILDTIME

RUN apk add --no-cache git make build-base

ENV SERVICE=omegabrr

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

#ENV GOOS=linux
#ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${REVISION} -X main.date=${BUILDTIME}" -o bin/example cmd/example/main.go

# build runner
FROM alpine:latest

LABEL org.opencontainers.image.source = "https://github.com/autobrr/omegabrr"

RUN useradd -r -u 1001 -g omegabrr omegabrr
USER omegabrr

ENV HOME="/config" \
XDG_CONFIG_HOME="/config" \
XDG_DATA_HOME="/config"

RUN apk --no-cache add ca-certificates curl

COPY --from=app-builder /src/bin/omegabrr /usr/local/bin/

WORKDIR /config

VOLUME /config

EXPOSE 7441

ENTRYPOINT ["/usr/local/bin/omegabrr", "run", "--config", "/config"]
#CMD ["--config", "/config"]
