FROM golang:alpine as build
WORKDIR /opt/app
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        make git gcc g++ linux-headers libgcc libstdc++
COPY . .
RUN make install

FROM alpine:latest
COPY --from=build /go/bin/wbd /usr/local/bin/
COPY --from=build /go/bin/wbcli /usr/local/bin/
