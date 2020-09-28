FROM golang:1.14-alpine as build
WORKDIR /opt/app
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        make git gcc g++ linux-headers libgcc libstdc++ bash moreutils
COPY . .
RUN make install

FROM alpine:latest
WORKDIR /opt/app
ARG CI_PIPELINE_ID
ARG CI_COMMIT_REF_NAME
ARG CI_COMMIT_SHA
RUN apk add --no-cache ca-certificates jq bash && \
    addgroup -g 1000 user && \
    adduser -D -u 1000 -h /home/user -s /bin/bash -g "user" -G user user
USER user
RUN echo "${CI_PIPELINE_ID}-${CI_COMMIT_REF_NAME}-${CI_COMMIT_SHA}" > version
COPY --from=build \
            /go/bin/dnode \
            /go/bin/dncli \
            /usr/local/bin/

STOPSIGNAL SIGTERM
