FROM --platform=$BUILDPLATFORM alpine:3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --no-cache --platform linux/amd64,linux/arm64 --build-arg HELM_VERSION=3.10.0 --build-arg PLUGIN_VERSION=0.3.0 -t alpine/helm-unittest:3.10.0-0.3.0 .
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG HELM_VERSION
ARG PLUGIN_VERSION

ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz"
ENV PLUGIN_URL="https://github.com/helm-unittest/helm-unittest/"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

# Ensure to have latest packages
RUN apk upgrade --no-cache && \
    apk add --no-cache --update bash ca-certificates curl git && \
    curl --proto "=https" -L ${HELM_BASE_URL}/${HELM_TAR_FILE} | tar xvz && \
    mv ${TARGETOS}-${TARGETARCH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} --version ${PLUGIN_VERSION} && \
    rm -rf ${TARGETOS}-${TARGETARCH} && \
    apk del curl git bash && \
    rm -f /var/cache/apk/* && \
    addgroup -g 1000 -S helmgroup && \
    adduser -u 1000 -S -G helmgroup helmuser

VOLUME ["/apps"] 

USER 1000:1000

WORKDIR /apps

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]