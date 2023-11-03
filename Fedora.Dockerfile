FROM fedora:40

# variable "HELM_VERSION" must be passed as docker environment variables during the image build
# docker build --no-cache --build-arg HELM_VERSION=3.13.0 -t fedora/helm-unittest:test -f Fedora.Dockerfile .

ADD ./plugin.yaml ~/plugin.yaml
ADD ./plugin-dbg.yaml ~/plugin-dbg.yaml
ADD ./install-binary.sh ~/install-binary.sh

ARG HELM_VERSION

ENV HELM_BASE_URL="https://get.helm.sh"
ENV HELM_TAR_FILE="helm-v${HELM_VERSION}-linux-amd64.tar.gz"
ENV PLUGIN_URL="~"
# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

RUN yum install -y git && \
    curl -L ${HELM_BASE_URL}/${HELM_TAR_FILE} |tar xvz && \
    mv linux-amd64/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    helm plugin install ${PLUGIN_URL} && \
    rm -rf linux-amd64 && \
    yum remove -y git && \
    rm -rf /var/cache/yum/* ;

RUN groupadd -r helmgroup \
    && useradd -r helmuser -g helmgroup

USER helmuser

WORKDIR /apps
VOLUME [ "/apps" ]

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]