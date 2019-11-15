GO111MODULE=on
export GO111MODULE

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
