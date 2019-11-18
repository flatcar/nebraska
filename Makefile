GO111MODULE=on
export GO111MODULE

SHELL = /bin/bash
VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "quay.io/flatcar"
DOCKER_IMAGE_NEBRASKA ?= "nebraska"
DOCKER_IMAGE_POSTGRES ?= "nebraska-postgres"

.PHONY: all
all: backend tools frontend

.PHONY: check
check:
	go test -p 1 ./...

container_id:
	set -e; \
	docker build \
		--file Dockerfile.postgres-test \
		--tag kinvolk/nebraska-postgres-test \
		.; \
	trap "rm -f container_id.tmp container_id" ERR; \
	docker run \
		--privileged \
		--detach \
		--publish 127.0.0.1:5432:5432 \
		kinvolk/nebraska-postgres-test \
		>container_id.tmp; \
	docker exec \
		$$(cat container_id.tmp) \
		/wait_for_db_ready.sh; \
	mv container_id.tmp container_id

.PHONY: check-backend-with-container
check-backend-with-container: container_id
	set -e; \
	trap "docker kill $$(cat container_id); docker rm $$(cat container_id); rm -f container_id" EXIT; \
	go test -p 1 ./...

.PHONY: frontend
frontend:
	cd frontend && npm install && npm run build

.PHONY: frontend-watch
frontend-watch:
	cd frontend && npx webpack --watch-poll 1000 --watch --config ./webpack.config.js --mode development

.PHONY: backend
backend: tools/go-bindata tools/golangci-lint
	PATH="$(abspath tools):$${PATH}" go generate ./...
	# this is to get nice error messages when something doesn't
	# build (both the project and the tests), golangci-lint's
	# output in this regard in unreadable.
	go build ./...
	NEBRASKA_SKIP_TESTS=1 go test ./... >/dev/null
	./tools/golangci-lint run --fix
	go mod tidy
	go build -o bin/nebraska ./cmd/nebraska

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
		-f Dockerfile.nebraska .

.PHONY: container-postgres
container-postgres:
	$(DOCKER_CMD) build \
		--no-cache \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):latest" \
		-f Dockerfile.postgres .

.PHONY: container
container: container-nebraska container-postgres

.PHONY: backend-ci
backend-ci: backend test-clean-work-tree-backend check-backend-with-container
