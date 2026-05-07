.PHONY: all build deps binary test doc

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

NOW = $(shell date -u '+%Y%m%d%I%M%S')

DOCKER := docker
GO := go
GO_ENV := $(shell $(GO) env GOOS GOARCH)
GOOS ?= $(word 1,$(GO_ENV))
GOARCH ?= $(word 2,$(GO_ENV))
GOFLAGS ?= $(GOFLAGS:)
ROOT_DIR := $(realpath .)

# GOOS/GOARCH of the build host, used to determine whether we're cross-compiling or not
BUILDER_GOOS_GOARCH="$(GOOS)_$(GOARCH)"

ifneq ($(GOOS), darwin)
	EXTLDFLAGS =	# EXTLDFLAGS = -extldflags "-lm -lstdc++ -static"
else
	EXTLDFLAGS =
endif

GO_LINKER_FLAGS ?= --ldflags \
	'$(EXTLDFLAGS) -s -w '

all: build

build: deps binary

deps:
	@echo "$(OK_COLOR)==> Fetching dependencies...$(NO_COLOR)"
	$(GO) mod download

binary:
	@echo "$(OK_COLOR)==> Building binary ($(GOOS)/$(GOARCH))...$(NO_COLOR)"
	$(GO) build -a $(GOFLAGS) $(GO_LINKER_FLAGS) -o bin/$(GOOS)_$(GOARCH)/scim

test: deps
	@echo "$(OK_COLOR)==> Running tests...$(NO_COLOR)"
	$(GO) test $(GOFLAGS) -race ./...

doc:
	mkdir -p /tmp/tmpgoroot/doc
	rm -rf /tmp/tmpgopath/src/github.com/imulab/go-scim
	mkdir -p /tmp/tmpgopath/src/github.com/imulab/go-scim
	tar -c --exclude='.git' --exclude='tmp' . | tar -x -C /tmp/tmpgopath/src/github.com/imulab/go-scim
	@echo "$(OK_COLOR)==> Open http://localhost:6060/pkg/github.com/imulab/go-scim$(NO_COLOR)"
	GOROOT=/tmp/tmpgoroot/ GOPATH=/tmp/tmpgopath/ godoc -http=localhost:6060

docker:
	@echo "$(OK_COLOR)==> Building image scim:latest$(NO_COLOR)"
	docker build -t scim:latest .

compose:
	@echo "$(OK_COLOR)==> Starting local stack$(NO_COLOR)"
	docker-compose up