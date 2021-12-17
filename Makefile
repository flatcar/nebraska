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

.PHONY: all
all: backend tools frontend

.PHONY: check
check:
	$(MAKE) -C backend $@

.PHONY: check-code-coverage
check-code-coverage:
	$(MAKE) -C backend $@

.PHONY: coverage.out
coverage.out:
	$(MAKE) -C backend $@

.PHONY: print-code-coverage
print-code-coverage:
	$(MAKE) -C backend $@

.PHONY: container_id
container_id:
	$(MAKE) -C backend $@

.PHONY: check-backend-with-container
check-backend-with-container:
	$(MAKE) -C backend $@

.PHONY: frontend
frontend:
	$(MAKE) -C frontend

.PHONY: frontend-watch
frontend-watch: run-frontend

run-frontend:
	$(MAKE) -C frontend run

.PHONY: frontend-install
frontend-install:
	$(MAKE) -C frontend install

.PHONY: frontend-install-ci
frontend-install-ci:
	$(MAKE) -C frontend install-ci

.PHONY: frontend-build
frontend-build:
	$(MAKE) -C frontend build

.PHONY: frontend-test
frontend-test:
	$(MAKE) -C frontend test

.PHONY: frontend-lint
frontend-lint:
	$(MAKE) -C frontend lint

.PHONY: frontend-tsc
frontend-tsc:
	$(MAKE) -C frontend tsc

.PHONY: i18n
i18n:
	$(MAKE) -C frontend $@

run:
	$(MAKE) -j 2 run-frontend run-backend

run-backend: backend-binary
	$(MAKE) -C backend run

.PHONY: backend
backend:
	$(MAKE) -C backend

.PHONY: backend-binary
backend-binary:
	$(MAKE) -C backend build

.PHONY: test-clean-work-tree-backend
test-clean-work-tree-backend:
	$(MAKE) -C backend $@

.PHONY: tools
tools:
	$(MAKE) -C backend $@

.PHONY: image
image:
	$(DOCKER_CMD) build \
		--no-cache \
		--build-arg NEBRASKA_VERSION=$(VERSION) \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NEBRASKA):latest" \
		-f Dockerfile .

.PHONY: container
container: image

.PHONY: backend-ci
backend-ci:
	$(MAKE) -C backend ci

.PHONY: run-generators
run-generators:
	$(MAKE) -C backend $@

.PHONY: build-backend-binary
build-backend-binary:
	$(MAKE) -C backend build

.PHONY: backend-code-checks
backend-code-checks:
	$(MAKE) -C backend code-checks

.PHONY: swagger-install
swagger-install:
	$(MAKE) -C backend tools/swag

.PHONY: swagger-init
swagger-init:
	$(MAKE) -C backend $@
