NAME=twigex
VERSION ?= 0.15.0
MIN_CLIENT_VERSION ?= 0.14.0
MIN_UPGRADABLE_VERSION ?= 0.14.0
PACKAGE_NAME=$(NAME)-$(VERSION)-linux-amd64
ENV=env GOOS=linux
EXECUTABLES = npm go tar
FILES := twigex frontend/dist i18n templates LICENSE NOTICE .env.example

DISTRO_VERSION ?= $(DISTRO_VERSION:)
BUILD_DATE = $(shell date -u)
BUILD_HASH ?= $(GITHUB_SHA)
BUILD_TIMESTAMP ?= $(shell date -u +%s)

ifeq ($(DISTRO_VERSION),)
	PACKAGE_NAME := $(NAME)-$(VERSION)-linux-amd64
else
    PACKAGE_NAME=$(NAME)-$(VERSION)-$(DISTRO_VERSION)-amd64
endif

LDFLAGS += -X "github.com/twigex/twigex/model.Version=$(VERSION)"
LDFLAGS += -X "github.com/twigex/twigex/model.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "github.com/twigex/twigex/model.BuildHash=$(BUILD_HASH)"
LDFLAGS += -X "github.com/twigex/twigex/model.BuildTimestamp=$(BUILD_TIMESTAMP)"
LDFLAGS += -X "github.com/twigex/twigex/model.MinClientVersion=$(MIN_CLIENT_VERSION)"
LDFLAGS += -X "github.com/twigex/twigex/model.MinUpgradableVersion=$(MIN_UPGRADABLE_VERSION)"
LDFLAGS += $(EXTRA_LDFLAGS)

.PHONY: all clean build main lint lint-frontend lint-go test-all test-frontend check-headers check-i18n check-tidy check-vuln sync-i18n tidy

all: clean build

build: generate compile package

generate:
	cd frontend && \
		npm install && npm run build && \
		cd ..

generate-testing:
	cd frontend && \
		npm install && npm run build:testing && \
		cd ..

test:
	go test -v ./...

test-db:
	@if [ -z "$$TEST_MYSQL_DSN" ]; then \
		echo "TEST_MYSQL_DSN is unset. Example:"; \
		echo "  export TEST_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/?multiStatements=true'"; \
		exit 1; \
	fi
	go test -v ./store/sqlstore/...

lint: lint-go lint-frontend

GOLANGCI_LINT_VERSION := v2.14.0

# golangci-lint's standard set includes go vet.
lint-go:
	@unformatted=$$(gofmt -l $$(find . -type f -name '*.go' \
		-not -path './frontend/node_modules/*' -not -path './vendor/*')); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed:"; echo "$$unformatted"; exit 1; \
	fi
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

lint-frontend:
	cd frontend && npm ci && npm run format:check && npm run lint

test-all:
	@if [ -z "$$TEST_MYSQL_DSN" ]; then \
		echo "TEST_MYSQL_DSN is unset; the store suite would skip silently."; \
		exit 1; \
	fi
	go test -race ./...

check-tidy:
	@mv enterprise/imports.go enterprise/imports.go.orig; \
	go mod tidy -diff; status=$$?; \
	mv enterprise/imports.go.orig enterprise/imports.go; \
	if [ $$status -ne 0 ]; then \
		echo "go.mod or go.sum is not tidy. Run 'make tidy' and commit the result."; \
		exit $$status; \
	fi

test-frontend:
	cd frontend && npm ci && npm test -- --run

GOVULNCHECK_VERSION := v1.8.0

check-vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) -format json ./... | go run ./tools/vulncheck

check-headers:
	@missing=$$(for f in $$(find . -type f \
		\( -name '*.go' -o -name '*.js' -o -name '*.css' -o -name '*.vue' -o -name '*.html' \) \
		-not -path './frontend/node_modules/*' -not -path './frontend/dist/*' \
		-not -path './vendor/*' -not -path './build/*'); do \
		head -5 "$$f" | grep -q SPDX-License-Identifier || echo "$$f"; \
	done); \
	if [ -n "$$missing" ]; then \
		echo "missing licence header:"; echo "$$missing"; exit 1; \
	fi

check-i18n:
	go run ./tools/i18n

sync-i18n:
	go run ./tools/sync-translations

tidy:
	mv enterprise/imports.go enterprise/imports.go.orig
	go mod tidy
	mv enterprise/imports.go.orig enterprise/imports.go

compile:
	go build -ldflags '$(LDFLAGS)'

package:
	mkdir -p build
	tar cvzf build/$(PACKAGE_NAME).tar.gz $(FILES)

docker: build
	docker build -t twigex .

start:
	./twigex start

clean:
	rm twigex
	rm frontend/package-lock.json
	rm -rf frontend/node_modules
	rm -rf frontend/dist
	rm -rf $(PACKAGE_NAME).tar.gz
