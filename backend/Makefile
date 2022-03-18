GO111MODULE=on
export GO111MODULE

TAG := `git describe --tags --always`
SHELL = /bin/bash
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "ghcr.io/kinvolk"
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

.PHONY: tools/codegen
tools/codegen:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0

.PHONY: codegen
codegen: tools/codegen
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=server --package codegen -o ./pkg/codegen/server.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=spec --package codegen -o ./pkg/codegen/spec.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=client --package codegen -o ./pkg/codegen/client.gen.go ./api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=types --package codegen -o ./pkg/codegen/types.gen.go ./api/spec.yaml;
