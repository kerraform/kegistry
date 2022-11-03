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
