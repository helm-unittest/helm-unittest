FROM --platform=$BUILDPLATFORM helmunittest/helm-unittest:1.1.2-fedora

# variable "HELM_VERSION" and "PLUGIN_VERSION" must be passed as docker environment variables during the image build
# docker buildx build --load --no-cache --platform linux/amd64 --build-arg HELM_VERSION=3.13.0 -t alpine/helm-unittest:test -f AlpineTest.Dockerfile .

ARG BUILDPLATFORM

COPY test/data/helmplugins /helmplugins

# Set user to install the plugin
USER root

# Ensure to have latest packages
RUN helm plugin install /helmplugins/annotation-tester

# Set user back to the non-root user
USER 1000:1000