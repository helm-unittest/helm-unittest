FROM --platform=$BUILDPLATFORM alpine@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

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
ENV OCI_PLUGIN_URL="oci://ghcr.io/helm-unittest/helm-unittest/unittest:${PLUGIN_VERSION}"

# Install the plugin for all users
ENV HELM_DATA_HOME=/usr/local/share/helm

# Copy plugin public key for verification
COPY public-key.asc /tmp/public-key.asc

# Ensure to have latest packages and handle version-specific installation
RUN apk upgrade --no-cache && \
    apk add --no-cache --update bash ca-certificates curl git gnupg && \
    gpg --import /tmp/public-key.asc && \
    gpg --export > ~/.gnupg/pubring.gpg && \
    curl --proto "=https" -L "${HELM_BASE_URL}/${HELM_TAR_FILE}" | tar xvz && \
    mv "${TARGETOS}-${TARGETARCH}/helm" /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    # Extract major version from HELM_VERSION (e.g., 4.0.0 -> 4) \
    HELM_MAJOR_VERSION=$(echo "${HELM_VERSION}" | cut -d. -f1) && \
    if [ $HELM_MAJOR_VERSION -ge 4 ]; then \
      if  [ "$(printf '%s\n%s' "1.1.0" "${PLUGIN_VERSION}" | sort -V | head -n1)" = "1.1.0" ];  then \
        helm plugin install "${OCI_PLUGIN_URL}"; \
      else \
        helm plugin install "${PLUGIN_URL}" --version "${PLUGIN_VERSION}" --verify=false; \
      fi \
    else \
      helm plugin install "${PLUGIN_URL}" --version "${PLUGIN_VERSION}"; \
    fi && \
    rm -rf "${TARGETOS}-${TARGETARCH}" && \
    apk del curl git bash && \
    rm -f /var/cache/apk/* && \
    addgroup -g 1000 -S helmgroup && \
    adduser -u 1000 -S -G helmgroup helmuser

VOLUME ["/apps"] 

USER 1000:1000

WORKDIR /apps

ENTRYPOINT ["helm", "unittest"]
CMD ["--help"]