DOCKERCMD = docker

GOCMD = go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

COMMIT = $(shell git rev-parse HEAD)
VERSION = unknown

.PHONY: build
build:
	@$(GOCMD) build \
		-ldflags '-X "github.com/kerraform/kegistry/internal/version.Version=$(VERSION)" -X "github.com/kerraform/kegistry/internal/version.Commit=$(COMMIT)"' \
		./main.go

.PHONY: run
run:
	@$(GOCMD) run \
		-ldflags '-X "github.com/kerraform/kegistry/internal/version.Version=$(VERSION)" -X "github.com/kerraform/kegistry/internal/version.Commit=$(COMMIT)"' \
		./main.go

.PHONY: run-jaeger
run-jaeger:
	@$(DOCKERCMD) run -d --name jaeger \
		-e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
		-p 5775:5775/udp \
		-p 6831:6831/udp \
		-p 6832:6832/udp \
		-p 5778:5778 \
		-p 16686:16686 \
		-p 14268:14268 \
		-p 9411:9411 \
		jaegertracing/all-in-one:1.6
