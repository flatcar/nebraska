VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_REPO ?= "schu"
DOCKER_IMAGE_ROLLERD ?= "coreroller-rollerd"
DOCKER_IMAGE_POSTGRES ?= "coreroller-postgres"

all: build

.PHONY: all

build:
	go build -o bin/rollerd cmd/rollerd/rollerd.go cmd/rollerd/controller.go
	go build -o bin/initdb cmd/initdb/initdb.go

.PHONY: build

container-rollerd:
	docker build \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):latest" \
		-f Dockerfile.rollerd .

.PHONY: container-rollerd

container-postgres:
	docker build \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):latest" \
		-f Dockerfile.postgres .

.PHONY: container-postgres

container: container-rollerd container-postgres

.PHONY: container
