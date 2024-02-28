# build app
FROM golang:1.21-alpine3.19 AS app-builder

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
ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w -X buildinfo.Version=${VERSION} -X buildinfo.Commit=${REVISION} -X buildinfo.Date=${BUILDTIME}" -o bin/omegabrr cmd/omegabrr/main.go

# build runner
FROM alpine:latest

LABEL org.opencontainers.image.source = "https://github.com/autobrr/omegabrr"

ENV APP_DIR="/app" CONFIG_DIR="/config" PUID="1000" PGID="1000" UMASK="002" TZ="Etc/UTC" ARGS=""
ENV XDG_CONFIG_HOME="${CONFIG_DIR}/.config" XDG_CACHE_HOME="${CONFIG_DIR}/.cache" XDG_DATA_HOME="${CONFIG_DIR}/.local/share" LANG="C.UTF-8" LC_ALL="C.UTF-8"

VOLUME ["${CONFIG_DIR}"]

# install packages
RUN apk add --no-cache tzdata shadow bash curl wget jq grep sed coreutils findutils unzip p7zip ca-certificates

COPY --from=app-builder /src/bin/omegabrr /usr/bin/

# make folders
RUN mkdir "${APP_DIR}" && \
# create user
    useradd -u 1000 -U -d "${CONFIG_DIR}" -s /bin/false omegabrr && \
    usermod -G users omegabrr

WORKDIR /config

EXPOSE 7441

ENTRYPOINT ["omegabrr", "run", "--config", "/config/config.yaml"]
#CMD ["--config", "/config"]
