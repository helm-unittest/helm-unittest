FROM golang:1.14.1-alpine3.11 as ALPINE-BUILDER
RUN apk --no-cache add --quiet alpine-sdk=1.0-r0
WORKDIR /go/src/github.com/lrills/helm-unittest/
COPY . .
RUN install -d /opt && make install HELM_PLUGIN_DIR=/opt

FROM alpine:3.11 as ALPINE
COPY --from=ALPINE-BUILDER /opt /opt
