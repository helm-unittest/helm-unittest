FROM --platform=$BUILDPLATFORM alpine:3@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t alpine/helm-unittest:test -f AlpineTest.Dockerfile .

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG HELM_VERSION

COPY plugin.yaml helm-unittest/plugin.yaml
COPY install-binary.sh helm-unittest/install-binary.sh
COPY untt helm-unittest/untt

ENV SKIP_BIN_INSTALL=1
ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz"
ENV PLUGIN_URL="helm-unittest"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

# Ensure to have latest packages
RUN apk upgrade --no-cache && \
    apk add --no-cache --update ca-certificates curl git libc6-compat && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv ${TARGETOS}-${TARGETARCH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf ${TARGETOS}-${TARGETARCH} && \
    apk del curl git && \
    rm -f /var/cache/apk/* && \
    addgroup -S helmgroup && \
    adduser -u 1000 -S helmuser -G helmgroup

USER helmuser

WORKDIR /apps

VOLUME ["/apps"] 

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]