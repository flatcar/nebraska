FROM golang:1.16-alpine as nebraska-build

ARG NEBRASKA_VERSION=""

ENV GOPATH=/go \
    GOPROXY=https://proxy.golang.org \
	GO111MODULE=on

# We optionally allow to set the version to display for the image.
# This is mainly used because when copying the source dir, docker will
# ignore the files we requested it to, and thus produce a "dirty" build
# as git status returns changes (when effectively for the built source
# there's none).
ENV VERSION=${NEBRASKA_VERSION}

RUN apk update && \
	apk add gcc git nodejs npm ca-certificates make musl-dev bash

COPY . /nebraska-source/

WORKDIR /nebraska-source

RUN make frontend build-backend-binary

FROM alpine:3.14.0

RUN apk update && \
	apk add ca-certificates tzdata

COPY --from=nebraska-build /nebraska-source/backend/bin/nebraska /nebraska/
COPY --from=nebraska-build /nebraska-source/frontend/build/ /nebraska/static/

ENV NEBRASKA_DB_URL "postgres://postgres@postgres:5432/nebraska?sslmode=disable&connect_timeout=10"
EXPOSE 8000
CMD ["/nebraska/nebraska", "-http-static-dir=/nebraska/static"]
