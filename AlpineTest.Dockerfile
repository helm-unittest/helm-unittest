FROM alpine:3

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker build --no-cache --build-arg HELM_VERSION=3.3.0 -t alpine/helm-unittest:test -f AlpineTest.Dockerfile .

ADD . ~

ARG HELM_VERSION

ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-linux-amd64.tar.gz"
ENV PLUGIN_URL="~"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

RUN if printf "$HELM_VERSION" | grep -Eq '^3.+'; then \
    apk add --update --no-cache curl ca-certificates git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf linux-amd64 && \
    apk del curl git && \
    rm -f /var/cache/apk/* ; \
else \
    apk add --update --no-cache curl ca-certificates git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm init --client-only && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf linux-amd64 && \
    apk del curl git && \
    rm -f /var/cache/apk/* ; \
fi

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]