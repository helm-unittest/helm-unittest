FROM golang:1.12.7-alpine3.9 as builder
RUN apk --no-cache add --quiet alpine-sdk=1.0-r0
WORKDIR /go/src/github.com/lrills/helm-unittest/
COPY . .
RUN install -d /opt && make install HELM_PLUGIN_DIR=/opt

FROM alpine:3.9
COPY --from=builder /opt /opt
