FROM centos:7

# variable "HELM_VERSION" must be passed as docker environment variables during the image build
# docker build --no-cache --build-arg HELM_VERSION=3.3.0 -t centos/helm-unittest:test -f CentOSTest.Dockerfile .

ADD . ~

ARG HELM_VERSION

ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-linux-amd64.tar.gz"
ENV PLUGIN_URL="~"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

RUN if printf "$HELM_VERSION" | grep -Eq '^3.+'; then \
    yum install -y git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf linux-amd64 && \
    yum remove -y git && \
    rm -rf /var/cache/yum/* ; \
else \
    yum install -y git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm init --client-only && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf linux-amd64 && \
    yum remove -y git && \
    rm -rf /var/cache/yum/* ; \
fi

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]