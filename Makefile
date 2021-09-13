GO111MODULE=on
export GO111MODULE

TAG := `git describe --tags --always`
SHELL = /bin/bash
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "ghcr.io/kinvolk"
DOCKER_IMAGE_NEBRASKA ?= "nebraska"
ifndef $(GOPATH)
	GOPATH=$(shell go env GOPATH)
	export GOPATH
endif

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

LDFLAGS := "-X github.com/kinvolk/nebraska/backend/pkg/version.Version=$(VERSION) -extldflags "-static""
.PHONY: all
all: backend tools frontend

.PHONY: check
check:
	cd backend && \
	go test -p 1 ./...
check-code-coverage:
	cd backend && \
	go test -p 1 -coverprofile=coverage.out ./...
coverage.out:
	make check-code-coverage
print-code-coverage: coverage.out
	cd backend && \
	go tool cover -html=coverage.out
container_id:
	cd backend && \
	./tools/setup_local_db.sh \
		--id-file container_id.tmp \
		--db-name nebraska_tests \
		--password nebraska
	cd backend && mv container_id.tmp container_id

.PHONY: check-backend-with-container
check-backend-with-container: container_id
	set -e; \
	cd backend && \
	trap "$(DOCKER_CMD) kill $$(cat container_id); $(DOCKER_CMD) rm $$(cat container_id); rm -f container_id" EXIT; \
	go test -p 1 ./...

.PHONY: frontend
frontend: frontend-install
	cd frontend && npm run build

.PHONY: frontend-watch
frontend-watch: run-frontend

run-frontend:
	cd frontend && npm start

.PHONY: frontend-install
frontend-install:
	cd frontend && npm install

.PHONY: frontend-install-ci
frontend-install-ci:
	cd frontend && npm ci

.PHONY: frontend-build
frontend-build:
	cd frontend && npm run build

.PHONY: frontend-test
frontend-test:
	cd frontend && npm run test

.PHONY: frontend-lint
frontend-lint:
	cd frontend && npm run lint

.PHONY: frontend-tsc
frontend-tsc:
	cd frontend && npm run tsc

.PHONY: i18n
i18n:
	cd frontend && npm run i18n

run-backend: backend-binary
	cd backend && ./bin/nebraska -auth-mode noop -debug

.PHONY: backend
backend: codegen run-generators backend-code-checks build-backend-binary

.PHONY: backend-binary
backend-binary: run-generators build-backend-binary

.PHONY: test-clean-work-tree-backend
test-clean-work-tree-backend:
	@cd backend && \
	if ! git diff --quiet -- go.mod go.sum pkg cmd tools/tools.go; then \
	  echo; \
	  echo 'Working tree of backend code is not clean'; \
	  echo; \
	  git status; \
	  exit 1; \
	fi

.PHONY: tools
tools:
	cd backend && go build -o bin/initdb ./cmd/initdb
	cd backend && go build -o bin/userctl ./cmd/userctl

backend/tools/go-bindata: backend/go.mod backend/go.sum
	cd backend && go build -o ./tools/go-bindata github.com/kevinburke/go-bindata/go-bindata

backend/tools/golangci-lint: backend/go.mod backend/go.sum
	cd backend && go build -o ./tools/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: image
image:
	$(DOCKER_CMD) build \
		--no-cache \
		--build-arg NEBRASKA_VERSION=$(VERSION) \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):latest" \
		-f Dockerfile .

.PHONY: backend/tools/codegen
backend/tools/codegen:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: codegen
codegen: backend/tools/codegen
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=server --package codegen -o ./backend/pkg/codegen/server.gen.go ./backend/api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=spec --package codegen -o ./backend/pkg/codegen/spec.gen.go ./backend/api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=client --package codegen -o ./backend/pkg/codegen/client.gen.go ./backend/api/spec.yaml;
	PATH=$$GOPATH/bin:$$PATH oapi-codegen --generate=types --package codegen -o ./backend/pkg/codegen/types.gen.go ./backend/api/spec.yaml;

.PHONY: container
container: image

.PHONY: backend-ci
backend-ci: backend test-clean-work-tree-backend check-backend-with-container

.PHONY: run-generators
run-generators: backend/tools/go-bindata
	cd backend && PATH="$(abspath backend/tools):$${PATH}" go generate ./...

.PHONY: build-backend-binary
build-backend-binary:
	cd backend && go build -trimpath -ldflags ${LDFLAGS} -o bin/nebraska ./cmd/nebraska

.PHONY: backend-code-checks
backend-code-checks: backend/tools/golangci-lint
	# this is to get nice error messages when something doesn't
	# build (both the project and the tests), golangci-lint's
	# output in this regard in unreadable.
	cd backend && go build ./...
	cd backend && ./tools/check_pkg_test.sh
	cd backend && NEBRASKA_SKIP_TESTS=1 go test ./... >/dev/null
	cd backend && ./tools/golangci-lint run --fix
	cd backend && go mod tidy

