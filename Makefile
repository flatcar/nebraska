VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_CMD ?= "docker"
DOCKER_REPO ?= "quay.io/flatcar"
DOCKER_IMAGE_ROLLERD ?= "nebraska-rollerd"
DOCKER_IMAGE_POSTGRES ?= "nebraska-postgres"

.PHONY: all
all: backend tools frontend

.PHONY: check
check:
	go test -p 1 ./...

.PHONY: frontend
frontend:
	cd frontend && npm install && npm run build

.PHONY: backend
backend:
	go build -o bin/rollerd ./cmd/rollerd

.PHONY: tools
tools:
	go build -o bin/initdb ./cmd/initdb
	go build -o bin/userctl ./cmd/userctl

.PHONY: container-rollerd
container-rollerd:
	$(DOCKER_CMD) build \
		--no-cache \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):latest" \
		-f Dockerfile.rollerd .

.PHONY: container-postgres
container-postgres:
	$(DOCKER_CMD) build \
		--no-cache \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):latest" \
		-f Dockerfile.postgres .

.PHONY: container
container: container-rollerd container-postgres
