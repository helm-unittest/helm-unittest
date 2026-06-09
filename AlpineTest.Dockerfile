FROM --platform=$BUILDPLATFORM alpine@sha256:4f4ba248d8a2c90a6e52ffdfc194181f7617f9ddaca348d4c550a6b354fc7c2a

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t alpine/helm-unittest:test -f AlpineTest.Dockerfile .

ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG HELM_VERSION

COPY plugin.yaml helm-unittest/plugin.yaml
COPY install-binary.sh helm-unittest/install-binary.sh
COPY untt helm-unittest/untt-${TARGETOS}-${TARGETARCH}

ENV SKIP_BIN_INSTALL=1
ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz"
ENV PLUGIN_URL="helm-unittest"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

# Ensure to have latest packages
RUN test -n "${TARGETOS}" && \
    test -n "${TARGETARCH}" && \
    apk upgrade --no-cache && \
    apk add --no-cache --update ca-certificates curl git libc6-compat yq && \
    curl --proto "=https" -L "${HELM_BASE_URL}/${HELM_TAR_FILE}" |tar xvz && \
    mv "${TARGETOS}-${TARGETARCH}/helm" /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install "${PLUGIN_URL}" && \
    rm -rf "${TARGETOS}-${TARGETARCH}" && \
    apk del curl git && \
    rm -f /var/cache/apk/* && \
    addgroup -g 1000 -S helmgroup && \
    adduser -u 1000 -S -G helmgroup helmuser

VOLUME ["/apps"]

USER 1000:1000

WORKDIR /apps

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]
