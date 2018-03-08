
# borrowed from https://github.com/technosophos/helm-template

HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-unittest
HAS_DEP := $(shell command -v dep;)
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION}"

.PHONY: install
install: bootstrap build
	cp unittest $(HELM_PLUGIN_DIR)
	cp plugin.yaml $(HELM_PLUGIN_DIR)

.PHONY: hookInstall
hookInstall: bootstrap build

.PHONY: build
build:
	go build -o untt -ldflags $(LDFLAGS) ./main.go

.PHONY: dist
dist:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-linux-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	GOOS=darwin GOARCH=amd64 go build -o untt -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-macos-$(VERSION).tgz untt README.md LICENSE plugin.yaml
	GOOS=windows GOARCH=amd64 go build -o untt.exe -ldflags $(LDFLAGS) ./main.go
	tar -zcvf $(DIST)/helm-unittest-windows-$(VERSION).tgz untt.exe README.md LICENSE plugin.yaml

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
	dep ensure
