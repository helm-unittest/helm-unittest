FROM --platform=$BUILDPLATFORM fedora:41@sha256:3ec60eb34fa1a095c0c34dd37cead9fd38afb62612d43892fcf1d3425c32bc1e

# variable "HELM_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t fedora/helm-unittest:test -f Fedora.Dockerfile .

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

RUN yum install -y git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv ${TARGETOS}-${TARGETARCH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf ${TARGETOS}-${TARGETARCH} && \
    yum remove -y git && \
    rm -rf /var/cache/yum/* && \
    groupadd -g 1000 -r helmgroup && \
    useradd -u 1000 -r -g helmgroup helmuser

VOLUME ["/apps"] 

USER 1000:1000

WORKDIR /apps

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]