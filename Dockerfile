FROM golang:1.12.4-alpine3.9 as ALPINE-BUILDER
RUN apk --no-cache add --quiet make curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/lrills/helm-unittest/
COPY . .
RUN install -d /opt && make install HELM_PLUGIN_DIR=/opt

FROM alpine:3.9 as ALPINE
COPY --from=ALPINE-BUILDER /opt /opt
