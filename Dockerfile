# syntax=docker/dockerfile:1.3

FROM docker.io/library/golang:1.16 as backend

ARG NEBRASKA_VERSION=""

ENV GOPATH=/go \
	GOPROXY=https://proxy.golang.org \
	GO111MODULE=on \
	CGO_ENABLED=0

# We optionally allow to set the version to display for the image.
# This is mainly used because when copying the source dir, docker will
# ignore the files we requested it to, and thus produce a "dirty" build
# as git status returns changes (when effectively for the built source
# there's none).
ENV VERSION=${NEBRASKA_VERSION}

WORKDIR /app

COPY backend .

RUN --mount=type=cache,target=$GOPATH/pkg/mod \
	make build

FROM docker.io/library/node:15 as frontend

WORKDIR /app

COPY frontend/package*.json .

RUN --mount=type=cache,target=node_modules \
	npm install

COPY frontend .

RUN --mount=type=cache,target=node_modules \
	make

FROM gcr.io/distroless/static

COPY --from=backend /app/bin/nebraska /nebraska/
COPY --from=frontend /app/build/ /nebraska/static/

ENV NEBRASKA_DB_URL "postgres://postgres@postgres:5432/nebraska?sslmode=disable&connect_timeout=10"
EXPOSE 8000
USER nobody
ENTRYPOINT ["/nebraska/nebraska", "-http-static-dir=/nebraska/static"]
