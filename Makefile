VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_REPO ?= "schu"
DOCKER_IMAGE_ROLLERD ?= "coreroller-rollerd"
DOCKER_IMAGE_POSTGRES ?= "coreroller-postgres"

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

.PHONY: container-rollerd
container-rollerd:
	docker build \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_ROLLERD):latest" \
		-f Dockerfile.rollerd .

.PHONY: container-postgres
container-postgres:
	docker build \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):$(VERSION)" \
		-t "$(DOCKER_REPO)/$(DOCKER_IMAGE_POSTGRES):latest" \
		-f Dockerfile.postgres .

.PHONY: container
container: container-rollerd container-postgres
