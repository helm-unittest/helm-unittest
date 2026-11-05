FROM --platform=$BUILDPLATFORM fedora@sha256:a7d72ba2745b100a3351087472e5ee7b2450af7cf331e7673007e3817612839c

# variable "HELM_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t fedora/helm-unittest:test -f FedoraTest.Dockerfile .

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

RUN test -n "${TARGETOS}" && \
    test -n "${TARGETARCH}" && \
    yum install -y git yq && \
    curl --fail --proto "=https" -L "${HELM_BASE_URL}/${HELM_TAR_FILE}" |tar xvz && \
    mv "${TARGETOS}-${TARGETARCH}/helm" /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install "${PLUGIN_URL}" && \
    rm -rf "${TARGETOS}-${TARGETARCH}" && \
    yum remove -y git && \
    rm -rf /var/cache/yum/* && \
    groupadd -g 1000 -r helmgroup && \
    useradd -m -u 1000 -r -g helmgroup helmuser

VOLUME ["/apps"] 

USER 1000:1000

WORKDIR /apps

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]