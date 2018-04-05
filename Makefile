APP = jamsonic

VERSION=$(shell \
        grep "const Version" version/version.go \
        |awk -F'=' '{print $$2}' \
        |sed -e "s/[^0-9.]//g" \
	|sed -e "s/ //g")

SHELL = /bin/bash

DIR = $(shell pwd)

GO = go

# GOX = gox -os="linux darwin windows freebsd openbsd netbsd"
# GOX = gox -os="linux"
# GOX_ARGS = "-output={{.Dir}}-$(VERSION)_{{.OS}}_{{.Arch}}"

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

MAKE_COLOR=\033[33;01m%-20s\033[0m

MAIN = github.com/TcM1911/jamsonic
EXE = $(shell ls jamsonic.exe)

OS=$(shell uname -s)
PACKAGE=$(APP)-$(VERSION)
ARCHIVE=$(PACKAGE)-$(OS).tar.gz
WINDOWS_ARCHIVE=$(APP)-$(VERSION)-Windows.zip
TAR_ARGS=cfvz
RELEASE_FILES = LICENSE README.md

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo -e "$(OK_COLOR)==== $(APP) [$(VERSION)] ====$(NO_COLOR)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(MAKE_COLOR) : %s\n", $$1, $$2}'

clean: ## Cleanup
	@echo -e "$(OK_COLOR)[$(APP)] Cleanup$(NO_COLOR)"
	@rm -fr $(APP) $(EXE) $(ARCHIVE) $(PACKAGE) $(WINDOWS_ARCHIVE) 2> /dev/null

.PHONY: init
init: ## Install requirements
	@echo -e "$(OK_COLOR)[$(APP)] Install requirements$(NO_COLOR)"
	@go get -u github.com/golang/glog
	@go get -u github.com/kardianos/govendor
	@go get -u github.com/Masterminds/rmvcsdir
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/kisielk/errcheck

.PHONY: deps
deps: ## Install dependencies
	@echo -e "$(OK_COLOR)[$(APP)] Update dependencies$(NO_COLOR)"
	@govendor update

.PHONY: build
build: ## Make binary
	@echo -e "$(OK_COLOR)[$(APP)] Build $(NO_COLOR)"
	@$(GO) build -o $(APP) -ldflags="-s -w" ./cmd/...

.PHONY: test
test: ## Launch unit tests
	@echo -e "$(OK_COLOR)[$(APP)] Launch unit tests $(NO_COLOR)"
	@govendor test +local

.PHONY: release
release: test build ## Make a release
	@echo -e "$(OK_COLOR)[$(APP)] Creating a release $(NO_COLOR)"
	mkdir -p $(PACKAGE) dist
	cp $(RELEASE_FILES) $(PACKAGE)
	cp $(APP) $(PACKAGE)
	tar $(TAR_ARGS) dist/$(ARCHIVE) $(PACKAGE)
	rm -r $(PACKAGE)

.PHONY: release_windows
release_windows:
	@echo -e "$(OK_COLOR)[$(APP)] Creating a release for Windows $(NO_COLOR)"
	GOOS=windows $(GO) build -o $(APP).exe -ldflags="-s -w" ./cmd/...
	mkdir -p $(PACKAGE) dist
	cp $(RELEASE_FILES) $(PACKAGE)
	cp $(APP).exe $(PACKAGE)
	zip -r dist/$(WINDOWS_ARCHIVE) $(PACKAGE)
	rm -r $(PACKAGE)

.PHONY: lint
lint: ## Launch golint
	@$(foreach file,$(SRCS),golint $(file) || exit;)

.PHONY: vet
vet: ## Launch go vet
	@$(foreach file,$(SRCS),$(GO) vet $(file) || exit;)

.PHONY: errcheck
errcheck: ## Launch go errcheck
	@echo -e "$(OK_COLOR)[$(APP)] Go Errcheck $(NO_COLOR)"
	@$(foreach pkg,$(PKGS),errcheck $(pkg) $(glide novendor) || exit;)

.PHONY: coverage
coverage: ## Launch code coverage
	@$(foreach pkg,$(PKGS),$(GO) test -cover $(pkg) $(glide novendor) || exit;)

# gox: ## Make all binaries
# 	@echo -e "$(OK_COLOR)[$(APP)] Create binaries $(NO_COLOR)"
# 	$(GOX) $(GOX_ARGS) github.com/TcM1911/jamsonic
