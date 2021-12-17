GO111MODULE=on
export GO111MODULE

TAG := `git describe --tags --always`
SHELL = /bin/bash
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "ghcr.io/kinvolk"
DOCKER_IMAGE_NEBRASKA ?= "nebraska"
VERSION ?=
ifeq ($(VERSION),)
	## Adds a '-dirty' suffix to version string if there are uncommitted changes
	changes := $(shell git status ./backend ./frontend --porcelain)
	ifeq ($(changes),)
		VERSION := $(TAG)
	else
		VERSION := $(TAG)-dirty
	endif
endif

LDFLAGS := "-w -X github.com/kinvolk/nebraska/backend/pkg/version.Version=$(VERSION) -extldflags \"-static\""
.PHONY: all
all: run-generators code-checks build

.PHONY: check
check:
	go test -p 1 ./...

coverage.out: check-code-coverage

check-code-coverage:
	go test -p 1 -coverprofile=coverage.out ./...

container_id:
	./tools/setup_local_db.sh \
		--id-file container_id.tmp \
		--db-name nebraska_tests \
		--password nebraska \
		--pg-version 13.3
	mv container_id.tmp container_id

.PHONY: check-backend-with-container
check-backend-with-container: container_id
	set -e; \
	trap "$(DOCKER_CMD) kill $$(cat container_id); $(DOCKER_CMD) rm $$(cat container_id); rm -f container_id" EXIT; \
	go test -p 1 ./...

run: bin/nebraska
	./bin/nebraska -auth-mode noop -debug

.PHONY: build
build: run-generators bin/nebraska

.PHONY: test-clean-work-tree-backend
test-clean-work-tree-backend:
	if ! git diff --quiet -- go.mod go.sum pkg cmd tools/tools.go; then \
	  echo; \
	  echo 'Working tree of backend code is not clean'; \
	  echo; \
	  git status; \
	  exit 1; \
	fi

.PHONY: tools
tools: bin/initdb bin/userctl

bin/initdb:
	go build -o bin/initdb ./cmd/initdb

bin/userctl:
	go build -o bin/userctl ./cmd/userctl

tools/go-bindata: go.mod go.sum
	env GOBIN=$(CURDIR)/tools/ go install github.com/kevinburke/go-bindata/go-bindata@v3.22.0

tools/golangci-lint: go.mod go.sum
	env GOBIN=$(CURDIR)/tools/ go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

tools/swag:
	env GOBIN=$(CURDIR)/tools/ go install github.com/swaggo/swag/cmd/swag@v1.7.4

.PHONY: ci
ci: build test-clean-work-tree-backend check-backend-with-container

.PHONY: run-generators
run-generators: tools/go-bindata
	PATH="$(abspath tools):$${PATH}" go generate ./...

bin/nebraska: run-generators
	 go build -a -tags netgo -trimpath -ldflags ${LDFLAGS} -o bin/nebraska ./cmd/nebraska

.PHONY: code-checks
code-checks: tools/golangci-lint
	# this is to get nice error messages when something doesn't
	# build (both the project and the tests), golangci-lint's
	# output in this regard in unreadable.
	go build ./...
	./tools/check_pkg_test.sh
	NEBRASKA_SKIP_TESTS=1 go test ./... >/dev/null
	./tools/golangci-lint run --fix
	go mod tidy

.PHONY: swagger-init
swagger-init:  tools/swag
	./tools/swag init -g cmd/userctl/main.go -o api
