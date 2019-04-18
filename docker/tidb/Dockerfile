# Builder image
FROM golang:1.12.4-alpine as builder

ENV TIDB_VERSION 2.1.8

RUN apk add --no-cache \
    wget \
    make \
    git \
    g++ \
    gcc

RUN git clone https://github.com/pingcap/tidb.git /go/src/github.com/pingcap/tidb

WORKDIR /go/src/github.com/pingcap/tidb/

RUN git checkout tags/v$TIDB_VERSION && rm go.sum && GO111MODULE=on go mod tidy && make

# Executable image
FROM alpine:3.9

COPY --from=builder /go/src/github.com/pingcap/tidb/bin/tidb-server /usr/local/bin/tidb-server

RUN apk add --no-cache \
    mysql-client \
    gnupg \
    openssl \
    pwgen \
    bash \
    tzdata

COPY /docker/tidb/entrypoint.sh /usr/local/bin/entrypoint.sh

WORKDIR /

EXPOSE 4000

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["/usr/local/bin/tidb-server"]
