SERVICE = omegabrr

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git tag --points-at HEAD 2> /dev/null | head -n 1)

GO ?= go
RM ?= rm
GOFLAGS ?= "-X github.com/autobrr/omegabrr/internal/buildinfo.Commit=$(GIT_COMMIT) -X github.com/autobrr/omegabrr/internal/buildinfo.Version=$(GIT_TAG)"
PREFIX ?= /usr/local
BINDIR ?= bin

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git tag --points-at HEAD 2> /dev/null | head -n 1)

all: clean build

deps:
	go mod download

build: deps
	go build -ldflags $(GOFLAGS) -o bin/$(SERVICE) cmd/$(SERVICE)/main.go

build/docker:
	docker build -t omegabrr:dev -f Dockerfile . --build-arg GIT_TAG=$(GIT_TAG) --build-arg GIT_COMMIT=$(GIT_COMMIT)

clean:
	$(RM) -rf bin

install: all
	echo $(DESTDIR)$(PREFIX)/$(BINDIR)
	mkdir -p $(DESTDIR)$(PREFIX)/$(BINDIR)
	cp -f bin/$(SERVICE) $(DESTDIR)$(PREFIX)/$(BINDIR)

.PHONY: build clean install