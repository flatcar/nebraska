FROM golang:1.23 AS base-build

ARG NEBRASKA_VERSION=""

ENV GOPATH=/go \
    GOPROXY=direct \
	GO111MODULE=on\
	CGO_ENABLED=0\ 
	GOOS=linux 

# Backend build
FROM base-build AS version-build
ARG NEBRASKA_VERSION=""

WORKDIR /app/
COPY ./.git ./.git
COPY ./backend ./backend

# We optionally allow to set the version to display for the image.
# This is mainly used because when copying the source dir, docker will
# ignore the files we requested it to, and thus produce a "dirty" build
# as git status returns changes (when effectively for the built source
# there's none).
ENV VERSION=${NEBRASKA_VERSION}

FROM version-build AS backend-build

# make version uses the existing VERSION if set, otherwise gets it from git
RUN export VERSION=`make -f backend/Makefile version | tail -1` && echo "VERSION:$VERSION"

WORKDIR /app/backend
# COPY backend/go.mod backend/go.sum ./
# RUN go mod download
COPY ./backend ./
RUN make build

# Frontend build
FROM docker.io/library/node:22 AS frontend-install
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install

FROM frontend-install AS frontend-build
WORKDIR /app/frontend
COPY frontend ./
RUN npm run build

# Final Docker image 
FROM alpine:3.21.3

RUN apk update && \
	apk add ca-certificates tzdata

WORKDIR /nebraska

COPY --from=backend-build /app/backend/bin/nebraska ./
COPY --from=frontend-build /app/frontend/dist/ ./static/

ENV NEBRASKA_DB_URL="postgres://postgres@postgres:5432/nebraska?sslmode=disable&connect_timeout=10"
EXPOSE 8000

USER nobody

CMD ["/nebraska/nebraska", "-http-static-dir=/nebraska/static"]
