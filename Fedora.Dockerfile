FROM --platform=$BUILDPLATFORM fedora:40

# variable "HELM_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t fedora/helm-unittest:test -f Fedora.Dockerfile .

ADD plugin.yaml helm-unittest/plugin.yaml
ADD install-binary.sh helm-unittest/install-binary.sh
ADD untt helm-unittest/untt

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG HELM_VERSION

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
    rm -rf /var/cache/yum/* ;

RUN groupadd -g 1000 -r helmgroup \
    && useradd -u 1000 -r helmuser -g helmgroup

USER helmuser

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]