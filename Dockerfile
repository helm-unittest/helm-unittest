FROM --platform=$BUILDPLATFORM alpine:3

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

RUN apk add --update --no-cache curl ca-certificates git bash && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv ${TARGETOS}-${TARGETARCH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} --version ${PLUGIN_VERSION} && \
    rm -rf ${TARGETOS}-${TARGETARCH} && \
    apk del curl git bash && \
    rm -f /var/cache/apk/*

RUN addgroup -S helmgroup && \
    adduser -S helmuser -G helmgroup

USER helmuser

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]