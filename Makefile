# Copyright © 2018 - 2023 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

export GO111MODULE=on

-include .env
export

# Binary file names.
BINARY_NAME=photoprism
GOIMPORTS=goimports

# Build parameters.
BUILD_PATH ?= $(shell realpath "./build")
BUILD_DATE ?= $(shell date -u +%y%m%d)
BUILD_VERSION ?= $(shell git describe --always)
BUILD_TAG ?= $(BUILD_DATE)-$(BUILD_VERSION)
BUILD_OS ?= $(shell uname -s)
BUILD_ARCH ?= $(shell scripts/dist/arch.sh)
JS_BUILD_PATH ?= $(shell realpath "./assets/static/build")

# Install parameters.
INSTALL_PATH ?= $(BUILD_PATH)/photoprism-ce-$(BUILD_TAG)-$(shell echo $(BUILD_OS) | tr '[:upper:]' '[:lower:]')-$(BUILD_ARCH)
DESTDIR ?= $(INSTALL_PATH)
DESTUID ?= 1000
DESTGID ?= 1000
INSTALL_USER ?= $(DESTUID):$(DESTGID)
INSTALL_MODE ?= u+rwX,a+rX
INSTALL_MODE_BIN ?= 755

UID := $(shell id -u)
GID := $(shell id -g)
HASRICHGO := $(shell which richgo)

ifdef HASRICHGO
    GOTEST=richgo test
else
    GOTEST=go test
endif

# Ensure compatibility with "docker-compose" (old) and "docker compose" (new).
HAS_DOCKER_COMPOSE_WITH_DASH := $(shell which docker-compose)

ifdef HAS_DOCKER_COMPOSE_WITH_DASH
    DOCKER_COMPOSE=docker-compose
else
    DOCKER_COMPOSE=docker compose
endif

# Declare "make" targets.
all: dep build-js
dep: dep-tensorflow dep-js
biuld: build
build: build-go
build-all: build-go build-js
pull: docker-pull
test: test-js test-go
test-go: reset-sqlite run-test-go
test-pkg: reset-sqlite run-test-pkg
test-api: reset-sqlite run-test-api
test-commands: reset-sqlite run-test-commands
test-photoprism: reset-sqlite run-test-photoprism
test-short: reset-sqlite run-test-short
test-mariadb: reset-acceptance run-test-mariadb
acceptance-run-chromium: storage/acceptance acceptance-auth-sqlite-restart wait acceptance-auth acceptance-auth-sqlite-stop acceptance-sqlite-restart wait-2 acceptance acceptance-sqlite-stop
acceptance-run-chromium-short: storage/acceptance acceptance-auth-sqlite-restart wait acceptance-auth-short acceptance-auth-sqlite-stop acceptance-sqlite-restart wait-2 acceptance-short acceptance-sqlite-stop
acceptance-auth-run-chromium: storage/acceptance acceptance-auth-sqlite-restart wait acceptance-auth acceptance-auth-sqlite-stop
acceptance-public-run-chromium: storage/acceptance acceptance-sqlite-restart wait acceptance acceptance-sqlite-stop
wait:
	sleep 20
wait-2:
	sleep 20
show-build:
	@echo "$(BUILD_TAG)"
test-all: test acceptance-run-chromium
fmt: fmt-js fmt-go
clean-local: clean-local-config clean-local-cache
upgrade: dep-upgrade-js dep-upgrade
devtools: install-go dep-npm
.SILENT: help;
logs:
	$(DOCKER_COMPOSE) logs -f
help:
	@echo "For build instructions, visit <https://docs.photoprism.app/developer-guide/>."
fix-permissions:
	$(info Updating filesystem permissions...)
	@if [ $(UID) != 0 ]; then\
		echo "Running \"chown --preserve-root -Rcf $(UID):$(GID) /go /photoprism /opt/photoprism /tmp/photoprism\". Please wait."; \
		sudo chown --preserve-root -Rcf $(UID):$(GID) /go /photoprism /opt/photoprism /tmp/photoprism || true;\
		echo "Running \"chmod --preserve-root -Rcf u+rwX /go/src/github.com/photoprism/* /photoprism /opt/photoprism /tmp/photoprism\". Please wait.";\
		sudo chmod --preserve-root -Rcf u+rwX /go/src/github.com/photoprism/photoprism/* /photoprism /opt/photoprism /tmp/photoprism || true;\
		echo "Done."; \
	else\
		echo "Running as root. Nothing to do."; \
	fi
gettext-merge:
	./scripts/gettext-merge.sh
gettext-clear-fuzzy:
	./scripts/gettext-clear-fuzzy.sh
clean:
	rm -f *.log .test*
	[ ! -f "$(BINARY_NAME)" ] || rm -f $(BINARY_NAME)
	[ ! -d "node_modules" ] || rm -rf node_modules
	[ ! -d "frontend/node_modules" ] || rm -rf frontend/node_modules
	[ ! -d "$(BUILD_PATH)" ] || rm -rf --preserve-root $(BUILD_PATH)
	[ ! -d "$(JS_BUILD_PATH)" ] || rm -rf --preserve-root $(JS_BUILD_PATH)
tar.gz:
	$(info Creating tar.gz archives from the directories in "$(BUILD_PATH)"...)
	find "$(BUILD_PATH)" -maxdepth 1 -mindepth 1 -type d -exec tar --exclude='.[^/]*' -C {} -czf {}.tar.gz . \;
pkg: pkg-amd64 pkg-arm64
pkg-amd64:
	docker run --rm -u $(UID) --platform=amd64 --pull=always -v ".:/go/src/github.com/photoprism/photoprism" --entrypoint "" photoprism/develop:jammy make all install tar.gz
pkg-arm64:
	docker run --rm -u $(UID) --platform=arm64 --pull=always -v ".:/go/src/github.com/photoprism/photoprism" --entrypoint "" photoprism/develop:jammy make all install tar.gz
install:
	$(info Installing in "$(DESTDIR)"...)
	@[ ! -d "$(DESTDIR)" ] || (echo "ERROR: Install path '$(DESTDIR)' already exists!"; exit 1)
	mkdir --mode=$(INSTALL_MODE) -p $(DESTDIR)
	env TMPDIR="$(BUILD_PATH)" ./scripts/dist/install-tensorflow.sh $(DESTDIR)
	rm -rf --preserve-root $(DESTDIR)/include
	(cd $(DESTDIR) && mkdir -p bin lib assets config config/examples)
	./scripts/build.sh prod "$(DESTDIR)/bin/$(BINARY_NAME)"
	rsync -r -l --safe-links --exclude-from=assets/.buildignore --chmod=a+r,u+rw ./assets/ $(DESTDIR)/assets
	wget -O $(DESTDIR)/assets/static/img/wallpaper/welcome.jpg https://cdn.photoprism.app/wallpaper/welcome.jpg
	wget -O $(DESTDIR)/assets/static/img/preview.jpg https://cdn.photoprism.app/img/preview.jpg
	cp internal/config/testdata/*.yml $(DESTDIR)/config/examples
	chown -R $(INSTALL_USER) $(DESTDIR)
	chmod -R $(INSTALL_MODE) $(DESTDIR)
	chmod -R $(INSTALL_MODE_BIN) $(DESTDIR)/bin $(DESTDIR)/lib
	@echo "PhotoPrism $(BUILD_TAG) has been successfully installed in \"$(DESTDIR)\".\nEnjoy!"
install-go:
	sudo scripts/dist/install-go.sh
	go build -v ./...
install-tensorflow:
	sudo scripts/dist/install-tensorflow.sh
install-darktable:
	sudo scripts/dist/install-darktable.sh
acceptance-sqlite-restart:
	cp -f storage/acceptance/backup.db storage/acceptance/index.db
	cp -f storage/acceptance/config-sqlite/settingsBackup.yml storage/acceptance/config-sqlite/settings.yml
	rm -rf storage/acceptance/sidecar/2020
	rm -rf storage/acceptance/sidecar/2011
	rm -rf storage/acceptance/originals/2010
	rm -rf storage/acceptance/originals/2020
	rm -rf storage/acceptance/originals/2011
	rm -rf storage/acceptance/originals/2013
	rm -rf storage/acceptance/originals/2017
	./photoprism --auth-mode="public" -c "./storage/acceptance/config-sqlite" --test start -d
acceptance-sqlite-stop:
	./photoprism --auth-mode="public" -c "./storage/acceptance/config-sqlite" --test stop
acceptance-auth-sqlite-restart:
	cp -f storage/acceptance/backup.db storage/acceptance/index.db
	cp -f storage/acceptance/config-sqlite/settingsBackup.yml storage/acceptance/config-sqlite/settings.yml
	./photoprism --auth-mode="password" -c "./storage/acceptance/config-sqlite" --test start -d
acceptance-auth-sqlite-stop:
	./photoprism --auth-mode="password" -c "./storage/acceptance/config-sqlite" --test stop
start:
	./photoprism start -d
stop:
	./photoprism stop
terminal:
	$(DOCKER_COMPOSE) exec -u $(UID) photoprism bash
rootshell: root-terminal
root-terminal:
	$(DOCKER_COMPOSE) exec -u root photoprism bash
migrate:
	go run cmd/photoprism/photoprism.go migrations run
generate:
	POT_SIZE_BEFORE=$(shell stat -L -c %s assets/locales/messages.pot)
	go generate ./pkg/... ./internal/...
	go fmt ./pkg/... ./internal/...
	POT_SIZE_AFTER=$(shell stat -L -c %s assets/locales/messages.pot)
	@if [ $(POT_SIZE_BEFORE) == $(POT_SIZE_AFTER) ]; then\
		git checkout -- assets/locales/messages.pot;\
		echo "Reverted unnecessary change in assets/locales/messages.pot.";\
	fi
go-generate:
	go generate ./pkg/... ./internal/...
	go fmt ./pkg/... ./internal/...
clean-local-assets:
	rm -rf $(BUILD_PATH)/assets/*
clean-local-cache:
	rm -rf $(BUILD_PATH)/storage/cache/*
clean-local-config:
	rm -f $(BUILD_PATH)/config/*
dep-list:
	go list -u -m -json all | go-mod-outdated -direct
dep-npm:
	sudo npm install -g npm
dep-js:
	(cd frontend && npm ci --no-update-notifier --no-audit)
dep-go:
	go build -v ./...
dep-upgrade:
	go get -u -t ./...
dep-upgrade-js:
	(cd frontend &&	npm --depth 3 update --legacy-peer-deps)
dep-tensorflow:
	scripts/download-facenet.sh
	scripts/download-nasnet.sh
	scripts/download-nsfw.sh
dep-acceptance: storage/acceptance
storage/acceptance:
	[ -f "./storage/acceptance/index.db" ] || (cd storage && rm -rf acceptance && wget -c https://dl.photoprism.app/qa/acceptance.tar.gz -O - | tar -xz)
zip-facenet:
	(cd assets && zip -r facenet.zip facenet -x "*/.*" -x "*/version.txt")
zip-nasnet:
	(cd assets && zip -r nasnet.zip nasnet -x "*/.*" -x "*/version.txt")
zip-nsfw:
	(cd assets && zip -r nsfw.zip nsfw -x "*/.*" -x "*/version.txt")
build-js:
	(cd frontend &&	env NODE_ENV=production npm run build)
build-go: build-debug
build-debug:
	rm -f $(BINARY_NAME)
	scripts/build.sh debug $(BINARY_NAME)
build-prod:
	rm -f $(BINARY_NAME)
	scripts/build.sh prod $(BINARY_NAME)
build-race:
	rm -f $(BINARY_NAME)
	scripts/build.sh race $(BINARY_NAME)
build-static:
	rm -f $(BINARY_NAME)
	scripts/build.sh static $(BINARY_NAME)
build-tensorflow:
	docker build -t photoprism/tensorflow:build docker/tensorflow
	docker run -ti photoprism/tensorflow:build bash
build-tensorflow-arm64:
	docker build -t photoprism/tensorflow:arm64 docker/tensorflow/arm64
	docker run -ti photoprism/tensorflow:arm64 bash
watch-js:
	(cd frontend &&	env NODE_ENV=development npm run watch)
test-js:
	$(info Running JS unit tests...)
	(cd frontend && env TZ=UTC NODE_ENV=development BABEL_ENV=test npm run test)
acceptance:
	$(info Running public-mode tests in 'chromium:headless'...)
	(cd frontend &&	npm run testcafe -- chrome:headless --test-grep "^(Common|Core)\:*" --test-meta mode=public --config-file ./testcaferc.json "tests/acceptance")
acceptance-short:
	$(info Running JS acceptance tests in Chrome...)
	(cd frontend &&	npm run testcafe -- chrome:headless --test-grep "^(Common|Core)\:*" --test-meta mode=public,type=short --config-file ./testcaferc.json "tests/acceptance")
acceptance-firefox:
	$(info Running JS acceptance tests in Firefox...)
	(cd frontend &&	npm run testcafe -- firefox:headless --test-grep "^(Common|Core)\:*" --test-meta mode=public --config-file ./testcaferc.json "tests/acceptance")
acceptance-auth:
	$(info Running JS acceptance-auth tests in Chrome...)
	(cd frontend &&	npm run testcafe -- chrome:headless --test-grep "^(Common|Core)\:*" --test-meta mode=auth --config-file ./testcaferc.json "tests/acceptance")
acceptance-auth-short:
	$(info Running JS acceptance-auth tests in Chrome...)
	(cd frontend &&	npm run testcafe -- chrome:headless --test-grep "^(Common|Core)\:*" --test-meta mode=auth,type=short --config-file ./testcaferc.json "tests/acceptance")
acceptance-auth-firefox:
	$(info Running JS acceptance-auth tests in Firefox...)
	(cd frontend &&	npm run testcafe -- firefox:headless --test-grep "^(Common|Core)\:*" --test-meta mode=auth --config-file ./testcaferc.json "tests/acceptance")
reset-mariadb:
	$(info Resetting photoprism database...)
	mysql < scripts/sql/reset-photoprism.sql
reset-mariadb-testdb:
	$(info Resetting testdb database...)
	mysql < scripts/sql/reset-testdb.sql
reset-mariadb-local:
	$(info Resetting local database...)
	mysql < scripts/sql/reset-local.sql
reset-mariadb-acceptance:
	$(info Resetting acceptance database...)
	mysql < scripts/sql/reset-acceptance.sql
reset-mariadb-all: reset-mariadb-testdb reset-mariadb-local reset-mariadb-acceptance reset-mariadb-photoprism
reset-testdb: reset-sqlite reset-mariadb-testdb
reset-acceptance: reset-mariadb-acceptance
reset-sqlite:
	$(info Removing test database files...)
	find ./internal -type f -name ".test.*" -delete
run-test-short:
	$(info Running short Go tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -short -timeout 5m ./pkg/... ./internal/...
run-test-go:
	$(info Running all Go tests...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
run-test-mariadb:
	$(info Running all Go tests on MariaDB...)
	PHOTOPRISM_TEST_DRIVER="mysql" PHOTOPRISM_TEST_DSN="root:photoprism@tcp(mariadb:4001)/acceptance?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true" $(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
run-test-pkg:
	$(info Running all Go tests in "/pkg"...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./pkg/...
run-test-api:
	$(info Running all API tests...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./internal/api/...
run-test-commands:
	$(info Running all CLI command tests...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./internal/commands/...
run-test-photoprism:
	$(info Running all Go tests in "/internal/photoprism"...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./internal/photoprism/...
test-parallel:
	$(info Running all Go tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./pkg/... ./internal/...
test-verbose:
	$(info Running all Go tests in verbose mode...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m -v ./pkg/... ./internal/...
test-race:
	$(info Running all Go tests with race detection in verbose mode...)
	$(GOTEST) -tags slow -race -timeout 60m -v ./pkg/... ./internal/...
test-coverage:
	$(info Running all Go tests with code coverage report...)
	go test -parallel 1 -count 1 -cpu 1 -failfast -tags slow -timeout 30m -coverprofile coverage.txt -covermode atomic ./pkg/... ./internal/...
	go tool cover -html=coverage.txt -o coverage.html
	go tool cover -func coverage.txt  | grep total:
docker-pull:
	$(DOCKER_COMPOSE) pull --ignore-pull-failures
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml pull --ignore-pull-failures
docker-build:
	$(DOCKER_COMPOSE) pull --ignore-pull-failures
	$(DOCKER_COMPOSE) build
docker-local-up:
	$(DOCKER_COMPOSE) -f docker-compose.local.yml up --force-recreate
docker-local-down:
	$(DOCKER_COMPOSE) -f docker-compose.local.yml down -V
develop: docker-develop
docker-develop: docker-develop-latest
docker-develop-all: docker-develop-latest docker-develop-other
docker-develop-latest: docker-develop-ubuntu
docker-develop-debian: docker-develop-bookworm docker-develop-bookworm-slim
docker-develop-ubuntu: docker-develop-mantic docker-develop-mantic-slim
docker-develop-other: docker-develop-debian docker-develop-bullseye docker-develop-bullseye-slim docker-develop-buster
docker-develop-bookworm:
	docker pull --platform=amd64 debian:bookworm-slim
	docker pull --platform=arm64 debian:bookworm-slim
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bookworm /bookworm "-t photoprism/develop:debian"
docker-develop-bookworm-slim:
	docker pull --platform=amd64 debian:bookworm-slim
	docker pull --platform=arm64 debian:bookworm-slim
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bookworm-slim /bookworm-slim
docker-develop-bullseye:
	docker pull --platform=amd64 golang:1-bullseye
	docker pull --platform=arm64 golang:1-bullseye
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bullseye /bullseye
docker-develop-bullseye-slim:
	docker pull --platform=amd64 debian:bullseye-slim
	docker pull --platform=arm64 debian:bullseye-slim
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bullseye-slim /bullseye-slim
develop-armv7: docker-develop-armv7
docker-develop-armv7:
	docker pull --platform=arm ubuntu:mantic
	scripts/docker/buildx.sh develop linux/arm armv7 /armv7
docker-develop-buster:
	docker pull --platform=amd64 golang:1-buster
	docker pull --platform=arm64 golang:1-buster
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 buster /buster
docker-develop-impish:
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 impish /impish
docker-develop-jammy:
	docker pull --platform=amd64 ubuntu:jammy
	docker pull --platform=arm64 ubuntu:jammy
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 jammy /jammy
docker-develop-jammy-slim:
	docker pull --platform=amd64 ubuntu:jammy
	docker pull --platform=arm64 ubuntu:jammy
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 jammy-slim /jammy-slim
docker-develop-lunar:
	docker pull --platform=amd64 ubuntu:lunar
	docker pull --platform=arm64 ubuntu:lunar
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 lunar /lunar
docker-develop-lunar-slim:
	docker pull --platform=amd64 ubuntu:lunar
	docker pull --platform=arm64 ubuntu:lunar
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 lunar-slim /lunar-slim
docker-develop-mantic:
	docker pull --platform=amd64 ubuntu:mantic
	docker pull --platform=arm64 ubuntu:mantic
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 mantic /mantic "-t photoprism/develop:latest -t photoprism/develop:ubuntu"
docker-develop-mantic-slim:
	docker pull --platform=amd64 ubuntu:mantic
	docker pull --platform=arm64 ubuntu:mantic
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 mantic-slim /mantic-slim
unstable: docker-unstable
docker-unstable: docker-unstable-mantic
docker-unstable-jammy:
	docker pull --platform=amd64 photoprism/develop:jammy
	docker pull --platform=amd64 photoprism/develop:jammy-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64 unstable-ce /jammy
docker-unstable-lunar:
	docker pull --platform=amd64 photoprism/develop:lunar
	docker pull --platform=amd64 photoprism/develop:lunar-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64 unstable-ce /lunar
docker-unstable-mantic:
	docker pull --platform=amd64 photoprism/develop:mantic
	docker pull --platform=amd64 photoprism/develop:mantic-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64 unstable-ce /mantic
preview: docker-preview-ce
docker-preview: docker-preview-ce
docker-preview-all: docker-preview-latest docker-preview-other
docker-preview-ce: docker-preview-mantic
docker-preview-latest: docker-preview-ubuntu
docker-preview-debian: docker-preview-bookworm
docker-preview-ubuntu: docker-preview-mantic
docker-preview-other: docker-preview-debian docker-preview-bullseye
docker-preview-arm: docker-preview-arm64 docker-preview-armv7
docker-preview-bookworm:
	docker pull --platform=amd64 photoprism/develop:bookworm
	docker pull --platform=amd64 photoprism/develop:bookworm-slim
	docker pull --platform=arm64 photoprism/develop:bookworm
	docker pull --platform=arm64 photoprism/develop:bookworm-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-bookworm /bookworm "-t photoprism/photoprism:preview-ce-debian"
docker-preview-armv7:
	docker pull --platform=arm photoprism/develop:armv7
	docker pull --platform=arm ubuntu:mantic
	scripts/docker/buildx.sh photoprism linux/arm preview-armv7 /armv7
docker-preview-arm64:
	docker pull --platform=arm64 photoprism/develop:lunar
	docker pull --platform=arm64 photoprism/develop:lunar-slim
	scripts/docker/buildx.sh photoprism linux/arm64 preview-arm64 /lunar
docker-preview-bullseye:
	docker pull --platform=amd64 photoprism/develop:bullseye
	docker pull --platform=amd64 photoprism/develop:bullseye-slim
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=arm64 photoprism/develop:bullseye-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-bullseye /bullseye
docker-preview-buster:
	docker pull --platform=amd64 photoprism/develop:buster
	docker pull --platform=arm64 photoprism/develop:buster
	docker pull --platform=amd64 debian:buster-slim
	docker pull --platform=arm64 debian:buster-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-buster /buster
docker-preview-impish:
	docker pull --platform=amd64 photoprism/develop:impish
	docker pull --platform=arm64 photoprism/develop:impish
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-impish /impish
docker-preview-jammy:
	docker pull --platform=amd64 photoprism/develop:jammy
	docker pull --platform=amd64 photoprism/develop:jammy-slim
	docker pull --platform=arm64 photoprism/develop:jammy
	docker pull --platform=arm64 photoprism/develop:jammy-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-ce /jammy
docker-preview-lunar:
	docker pull --platform=amd64 photoprism/develop:lunar
	docker pull --platform=amd64 photoprism/develop:lunar-slim
	docker pull --platform=arm64 photoprism/develop:lunar
	docker pull --platform=arm64 photoprism/develop:lunar-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-ce /lunar
docker-preview-mantic:
	docker pull --platform=amd64 photoprism/develop:mantic
	docker pull --platform=amd64 photoprism/develop:mantic-slim
	docker pull --platform=arm64 photoprism/develop:mantic
	docker pull --platform=arm64 photoprism/develop:mantic-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-ce /mantic
release: docker-release
docker-release: docker-release-latest
docker-release-all: docker-release-latest docker-release-other
docker-release-latest: docker-release-ubuntu
docker-release-debian: docker-release-bookworm
docker-release-ubuntu: docker-release-mantic
docker-release-other: docker-release-debian docker-release-bullseye
docker-release-arm: docker-release-arm64 docker-release-armv7
docker-release-bookworm:
	docker pull --platform=amd64 photoprism/develop:bookworm
	docker pull --platform=amd64 photoprism/develop:bookworm-slim
	docker pull --platform=arm64 photoprism/develop:bookworm
	docker pull --platform=arm64 photoprism/develop:bookworm-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce-bookworm /bookworm "-t photoprism/photoprism:ce-debian"
docker-release-armv7:
	docker pull --platform=arm photoprism/develop:armv7
	docker pull --platform=arm ubuntu:mantic
	scripts/docker/buildx.sh photoprism linux/arm armv7 /armv7
docker-release-arm64:
	docker pull --platform=arm64 photoprism/develop:lunar
	docker pull --platform=arm64 photoprism/develop:lunar-slim
	scripts/docker/buildx.sh photoprism linux/arm64 ce-arm64 /lunar
docker-release-bullseye:
	docker pull --platform=amd64 photoprism/develop:bullseye
	docker pull --platform=amd64 photoprism/develop:bullseye-slim
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=arm64 photoprism/develop:bullseye-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce-bullseye /bullseye
docker-release-buster:
	docker pull --platform=amd64 photoprism/develop:buster
	docker pull --platform=arm64 photoprism/develop:buster
	docker pull --platform=amd64 debian:buster-slim
	docker pull --platform=arm64 debian:buster-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce-buster /buster
docker-release-impish:
	docker pull --platform=amd64 photoprism/develop:impish
	docker pull --platform=arm64 photoprism/develop:impish
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce-impish /impish
docker-release-jammy:
	docker pull --platform=amd64 photoprism/develop:jammy
	docker pull --platform=amd64 photoprism/develop:jammy-slim
	docker pull --platform=arm64 photoprism/develop:jammy
	docker pull --platform=arm64 photoprism/develop:jammy-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce-jammy /jammy
docker-release-lunar:
	docker pull --platform=amd64 photoprism/develop:lunar
	docker pull --platform=amd64 photoprism/develop:lunar-slim
	docker pull --platform=arm64 photoprism/develop:lunar
	docker pull --platform=arm64 photoprism/develop:lunar-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce /lunar
docker-release-mantic:
	docker pull --platform=amd64 photoprism/develop:mantic
	docker pull --platform=amd64 photoprism/develop:mantic-slim
	docker pull --platform=arm64 photoprism/develop:mantic
	docker pull --platform=arm64 photoprism/develop:mantic-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 ce /mantic
start-local:
	$(DOCKER_COMPOSE) -f docker-compose.local.yml up -d --wait
stop-local:
	$(DOCKER_COMPOSE) -f docker-compose.local.yml stop
mysql:
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml pull mysql
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml stop mysql
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml up -d --wait mysql
start-mysql:
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml up -d --wait mysql
stop-mysql:
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml stop mysql
logs-mysql:
	$(DOCKER_COMPOSE) -f docker-compose.mysql.yml logs -f mysql
latest:
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml pull photoprism-latest
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml stop photoprism-latest
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml up -d --wait photoprism-latest
start-latest:
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml up photoprism-latest
stop-latest:
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml stop photoprism-latest
terminal-latest:
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml exec photoprism-latest bash
logs-latest:
	$(DOCKER_COMPOSE) -f docker-compose.latest.yml logs -f photoprism-latest
docker-local: docker-local-mantic
docker-local-all: docker-local-mantic docker-local-lunar docker-local-jammy docker-local-bookworm docker-local-bullseye docker-local-buster
docker-local-bookworm:
	docker pull photoprism/develop:bookworm
	docker pull photoprism/develop:bookworm-slim
	scripts/docker/build.sh photoprism ce-bookworm /bookworm "-t photoprism/photoprism:local"
docker-local-bullseye:
	docker pull photoprism/develop:bullseye
	docker pull photoprism/develop:bullseye-slim
	scripts/docker/build.sh photoprism ce-bullseye /bullseye "-t photoprism/photoprism:local"
docker-local-buster:
	docker pull photoprism/develop:buster
	docker pull debian:buster-slim
	scripts/docker/build.sh photoprism ce-buster /buster "-t photoprism/photoprism:local"
docker-local-impish:
	docker pull photoprism/develop:impish
	docker pull ubuntu:impish
	scripts/docker/build.sh photoprism ce-impish /impish "-t photoprism/photoprism:local"
docker-local-jammy:
	docker pull photoprism/develop:jammy
	docker pull ubuntu:jammy
	scripts/docker/build.sh photoprism ce-jammy /jammy "-t photoprism/photoprism:local"
docker-local-lunar:
	docker pull photoprism/develop:lunar
	docker pull ubuntu:lunar
	scripts/docker/build.sh photoprism ce-lunar /lunar "-t photoprism/photoprism:local"
docker-local-mantic:
	docker pull photoprism/develop:mantic
	docker pull ubuntu:mantic
	scripts/docker/build.sh photoprism ce-mantic /mantic "-t photoprism/photoprism:local"
docker-local-develop: docker-local-develop-mantic
docker-local-develop-all: docker-local-develop-mantic docker-local-develop-lunar docker-local-develop-jammy docker-local-develop-bookworm docker-local-develop-bullseye docker-local-develop-buster docker-local-develop-impish
docker-local-develop-bookworm:
	docker pull debian:bookworm-slim
	scripts/docker/build.sh develop bookworm /bookworm
docker-local-develop-bullseye:
	docker pull golang:1-bullseye
	scripts/docker/build.sh develop bullseye /bullseye
docker-local-develop-buster:
	docker pull golang:1-buster
	scripts/docker/build.sh develop buster /buster
docker-local-develop-impish:
	docker pull ubuntu:impish
	scripts/docker/build.sh develop impish /impish
docker-local-develop-jammy:
	docker pull ubuntu:jammy
	scripts/docker/build.sh develop jammy /jammy
docker-local-develop-lunar:
	docker pull ubuntu:lunar
	scripts/docker/build.sh develop lunar /lunar
docker-local-develop-mantic:
	docker pull ubuntu:mantic
	scripts/docker/build.sh develop mantic /mantic
docker-ddns:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh ddns linux/amd64,linux/arm64 $(BUILD_DATE)
docker-goproxy:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh goproxy linux/amd64,linux/arm64 $(BUILD_DATE)
demo: docker-demo
docker-demo: docker-demo-latest
docker-demo-all: docker-demo-latest docker-demo-debian
docker-demo-latest:
	docker pull photoprism/photoprism:preview-ce
	scripts/docker/build.sh demo ce
	scripts/docker/push.sh demo ce
docker-demo-debian:
	docker pull photoprism/photoprism:preview-ce-debian
	scripts/docker/build.sh demo debian /debian
	scripts/docker/push.sh demo debian
docker-demo-ubuntu:
	docker pull photoprism/photoprism:preview-ce-ubuntu
	scripts/docker/build.sh demo ubuntu /ubuntu
	scripts/docker/push.sh demo ubuntu
docker-demo-unstable:
	docker pull photoprism/photoprism:unstable-ce
	scripts/docker/build.sh demo $(BUILD_DATE) /unstable
	scripts/docker/push.sh demo $(BUILD_DATE)
docker-demo-local:
	scripts/docker/build.sh photoprism
	scripts/docker/build.sh demo $(BUILD_DATE) /debian
	scripts/docker/push.sh demo $(BUILD_DATE)
docker-dummy-webdav:
	docker pull --platform=amd64 golang:1
	docker pull --platform=arm64 golang:1
	scripts/docker/buildx-multi.sh dummy-webdav linux/amd64,linux/arm64 $(BUILD_DATE)
docker-dummy-oidc:
	docker pull --platform=amd64 golang:1
	docker pull --platform=arm64 golang:1
	scripts/docker/buildx-multi.sh dummy-oidc linux/amd64,linux/arm64 $(BUILD_DATE)
packer-digitalocean:
	$(info Buildinng DigitalOcean marketplace image...)
	(cd ./setup/docker/cloud && packer build digitalocean.json)
drone-sign:
	drone sign photoprism/photoprism --save
lint-js:
	(cd frontend &&	npm run lint)
fmt-js:
	(cd frontend &&	npm run fmt)
fmt-go:
	go fmt ./pkg/... ./internal/... ./cmd/...
	gofmt -w -s pkg internal cmd
	goimports -w pkg internal cmd
tidy:
	go mod tidy -go=1.16 && go mod tidy -go=1.17
users:
	./photoprism users add -p photoprism -r admin -s -a test:true -n "Alice Austen" superadmin
	./photoprism users ls

# Declare all targets as "PHONY", see https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html.
MAKEFLAGS += --always-make
.PHONY: all assets build cmd docker frontend internal pkg scripts storage photoprism install;
