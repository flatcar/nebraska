GO111MODULE=on
export GO111MODULE

BASE_TAG := $(shell git describe --tags --always --exclude 'nebraska-helm*' 2>/dev/null)

# We work only with one staging tag as per the one staging environment for Nebraska.
# Appending the short hash code helps checking the new deployments by
# asserting the version label in the footer that the hash changed.
TAG := $(shell \
    if [ "$(BASE_TAG)" = "staging" ]; then \
        echo "staging-$$(git rev-parse --short HEAD)"; \
    else \
        echo "$(BASE_TAG)"; \
    fi)

SHELL = /bin/bash
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "ghcr.io/flatcar"
DOCKER_IMAGE_NEBRASKA ?= "nebraska"
VERSION ?=
CONTAINER_NEBRASKA_DB_URL ?= "postgres://postgres:nebraska@127.0.0.1:8001/nebraska_tests?sslmode=disable&connect_timeout=10"
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
all: codegen run-generators code-checks build

.PHONY: check
check:
	go test -p 1 ./...

coverage.out: check-code-coverage

check-code-coverage:
	go test -p 1 -coverprofile=coverage.out ./...

.PHONY: test-service-up
test-service-up:
	docker compose -f ./docker-compose.test.yaml up -d --build

.PHONY: test-service-down
test-service-down:
	docker compose -f ./docker-compose.test.yaml down

.PHONY: check-backend-with-container
check-backend-with-container: test-service-down
	$(MAKE) test-service-up
	go clean -testcache && NEBRASKA_RUN_SERVER_TESTS=1 NEBRASKA_TEST_SERVER_URL="http://localhost:8002" NEBRASKA_DB_URL=$(CONTAINER_NEBRASKA_DB_URL) go test -p 1 ./...
	$(MAKE) test-service-down

run: bin/nebraska
	./bin/nebraska -auth-mode noop -debug

.PHONY: build
build: run-generators bin/nebraska

.PHONY: version
version:
	echo $(VERSION)

.PHONY: test-clean-work-tree-backend
test-clean-work-tree-backend:
	if ! git diff --quiet -- go.mod go.sum pkg cmd tools/tools.go; then \
	  echo; \
	  echo 'Working tree of backend code is not clean'; \
	  echo; \
	  git status; \
	  git diff; \
	  exit 1; \
	fi

.PHONY: tools
tools: bin/initdb

bin/initdb:
	go build -o bin/initdb ./cmd/initdb

tools/go-bindata: go.mod go.sum
	env GOBIN=$(CURDIR)/tools/ go install github.com/kevinburke/go-bindata/go-bindata@v3.24.0

tools/golangci-lint: go.mod go.sum
	env GOBIN=$(CURDIR)/tools/ go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6

tools/swag:
	env GOBIN=$(CURDIR)/tools/ go install github.com/swaggo/swag/cmd/swag@v1.8.10

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

.PHONY: tools/oapi-codegen
tools/oapi-codegen:
	env GOBIN=$(CURDIR)/tools/ go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0

.PHONY: codegen
codegen: tools/oapi-codegen
	PATH=$$GOPATH/bin:$$PATH ./tools/oapi-codegen --old-config-style --generate=server --package codegen -o ./pkg/codegen/server.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH ./tools/oapi-codegen --old-config-style --generate=spec --package codegen -o ./pkg/codegen/spec.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH ./tools/oapi-codegen --old-config-style --generate=client --package codegen -o ./pkg/codegen/client.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH ./tools/oapi-codegen --old-config-style --generate=types --package codegen -o ./pkg/codegen/types.gen.go ./api/spec.yaml;
