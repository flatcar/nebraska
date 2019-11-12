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
backend:
	go build -o bin/nebraska ./cmd/nebraska

.PHONY: tools
tools:
	go build -o bin/initdb ./cmd/initdb
	go build -o bin/userctl ./cmd/userctl

.PHONY: bindata
bindata:
	go generate ./pkg/api
	gofmt -s -w pkg/api/bindata.go

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
