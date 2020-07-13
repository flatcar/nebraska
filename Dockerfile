FROM alpine:3.12.0 as nebraska-build

ENV GOPATH=/go \
    GOPROXY=https://proxy.golang.org

RUN apk update && \
	apk add git go nodejs npm ca-certificates make musl-dev bash

COPY . /go/src/github.com/kinvolk/nebraska/

RUN cd /go/src/github.com/kinvolk/nebraska && \
	rm -rf frontend/node_modules tools/go-bindata tools/golangci-lint bin/nebraska && \
	make frontend backend-binary

FROM alpine:3.12.0

RUN apk update && \
	apk add ca-certificates tzdata

COPY --from=nebraska-build /go/src/github.com/kinvolk/nebraska/bin/nebraska /nebraska/
COPY --from=nebraska-build /go/src/github.com/kinvolk/nebraska/frontend/build/ /nebraska/static/
COPY --from=nebraska-build /go/src/github.com/kinvolk/nebraska/kinvolk_logo.svg /nebraska/assets/

ENV NEBRASKA_DB_URL "postgres://postgres@postgres:5432/nebraska?sslmode=disable&connect_timeout=10"
EXPOSE 8000
CMD ["/nebraska/nebraska", "-http-static-dir=/nebraska/static"]
