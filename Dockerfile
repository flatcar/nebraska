FROM alpine:3.12.0 as nebraska-build

ENV GOPATH=/go \
    GOPROXY=https://proxy.golang.org \
	GO111MODULE=on

RUN apk update && \
	apk add git go nodejs npm ca-certificates make musl-dev bash

COPY . /nebraska-source/

WORKDIR /nebraska-source

RUN make frontend backend-binary

FROM alpine:3.12.0

RUN apk update && \
	apk add ca-certificates tzdata

COPY --from=nebraska-build /nebraska-source/bin/nebraska /nebraska/
COPY --from=nebraska-build /nebraska-source/frontend/build/ /nebraska/static/

ENV NEBRASKA_DB_URL "postgres://postgres@postgres:5432/nebraska?sslmode=disable&connect_timeout=10"
EXPOSE 8000
CMD ["/nebraska/nebraska", "-http-static-dir=/nebraska/static"]
