GO111MODULE=on
export GO111MODULE

TAG := `git describe --tags --always`
VERSION :=
SHELL = /bin/bash
VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "quay.io/flatcar"
DOCKER_IMAGE_NEBRASKA ?= "nebraska"
## Adds a '-dirty' suffix to version string if there are uncommitted changes
changes := $(shell git status --porcelain)
ifeq ($(changes),)
	VERSION := $(TAG)
else
	VERSION := $(TAG)-dirty
endif

LDFLAGS := "-X github.com/kinvolk/nebraska/pkg/version.Version=$(VERSION) -extldflags "-static""
.PHONY: all
all: backend tools frontend

.PHONY: check
check:
	go test -p 1 ./...
check-code-coverage:
	go test -p 1 -coverprofile=coverage.out ./...
print-code-coverage: 
	go tool cover -html=coverage.out
container_id:
	set -e; \
	trap "rm -f container_id.tmp container_id" ERR; \
	docker run \
		--detach \
		--publish 127.0.0.1:5432:5432 \
		-e POSTGRES_PASSWORD=nebraska \
		postgres \
		>container_id.tmp; \
	until docker exec \
		$$(cat container_id.tmp) \
		pg_isready -h localhost; do sleep 3; done
	docker exec \
		$$(cat container_id.tmp) \
		psql -h localhost -U postgres -c 'create database nebraska_tests;'
	docker exec \
		$$(cat container_id.tmp) \
		psql -h localhost -U postgres -d nebraska_tests -c 'set timezone = "utc";'
	mv container_id.tmp container_id

.PHONY: check-backend-with-container
check-backend-with-container: container_id
	set -e; \
	trap "docker kill $$(cat container_id); docker rm $$(cat container_id); rm -f container_id" EXIT; \
	go test -p 1 ./...

.PHONY: frontend
frontend: frontend-install
	cd frontend && npm run build

.PHONY: frontend-watch
frontend-watch:
	cd frontend && npm start

.PHONY: frontend-install
frontend-install:
	cd frontend && npm install

.PHONY: frontend-test
frontend-test:
	cd frontend && npm run test

.PHONY: frontend-lint
frontend-lint:
	cd frontend && npm run lint

.PHONY: backend
backend: run-generators backend-code-checks build-backend-binary

.PHONY: backend-binary
backend-binary: run-generators build-backend-binary

.PHONY: test-clean-work-tree-backend
test-clean-work-tree-backend:
	@if ! git diff --quiet -- go.mod go.sum pkg cmd updaters tools/tools.go; then \
	  echo; \
	  echo 'Working tree of backend code is not clean'; \
	  echo; \
	  git status; \
	  exit 1; \
	fi

.PHONY: tools
tools:
	go build -o bin/initdb ./cmd/initdb
	go build -o bin/userctl ./cmd/userctl

tools/go-bindata: go.mod go.sum
	go build -o tools/go-bindata github.com/kevinburke/go-bindata/go-bindata

tools/golangci-lint: go.mod go.sum
	go build -o ./tools/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: container-nebraska
container-nebraska:
	$(DOCKER_CMD) build \
		--no-cache \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):latest" \
		-f Dockerfile .

.PHONY: container
container: container-nebraska

.PHONY: backend-ci
backend-ci: backend test-clean-work-tree-backend check-backend-with-container

.PHONY: run-generators
run-generators: tools/go-bindata
	PATH="$(abspath tools):$${PATH}" go generate ./...

.PHONY: build-backend-binary
build-backend-binary:
	go build -ldflags ${LDFLAGS} -o bin/nebraska ./cmd/nebraska

.PHONY: backend-code-checks
backend-code-checks: tools/golangci-lint
	# this is to get nice error messages when something doesn't
	# build (both the project and the tests), golangci-lint's
	# output in this regard in unreadable.
	go build ./...
	./tools/check_pkg_test.sh
	NEBRASKA_SKIP_TESTS=1 go test ./... >/dev/null
	./tools/golangci-lint run --fix
	go mod tidy
