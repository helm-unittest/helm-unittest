FROM alpine:3

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker build --no-cache --build-arg HELM_VERSION=3.10.0 PLUGIN_VERSION=0.3.0 -t alpine/helm-unittest:3.10.0-0.3.0 .

ARG HELM_VERSION
ARG PLUGIN_VERSION

ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-linux-amd64.tar.gz"
ENV PLUGIN_URL="https://github.com/quintush/helm-unittest/"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

RUN apk add --update --no-cache curl ca-certificates git bash && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} --version ${PLUGIN_VERSION} && \
    rm -rf linux-amd64 && \
    apk del curl git bash && \
    rm -f /var/cache/apk/*

RUN addgroup -S helmgroup && \
    adduser -S helmuser -G helmgroup

USER helmuser

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]