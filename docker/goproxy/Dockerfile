##################################################### BUILD STAGE ######################################################
FROM golang:alpine AS build

ARG GOPROXY
ARG GODEBUG

RUN apk add --no-cache -U make git mercurial subversion

RUN git clone https://github.com/goproxyio/goproxy.git /src/goproxy && \
    cd /src/goproxy && \
    export CGO_ENABLED=0 && \
    make

################################################## PRODUCTION STAGE ####################################################
FROM golang:alpine

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static-amd64 /usr/bin/tini
RUN chmod +x /usr/bin/tini

RUN apk add --no-cache -U git mercurial subversion

COPY --from=build /src/goproxy/bin/goproxy /goproxy
VOLUME "/go"

EXPOSE 8888

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/goproxy", "-listen", "0.0.0.0:8888"]