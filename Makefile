export GO111MODULE=on
export GODEBUG=http2client=0

GOIMPORTS=goimports
BINARY_NAME=photoprism

# Build Parameters
BUILD_PATH ?= $(shell realpath "./build")
BUILD_DATE ?= $(shell date -u +%y%m%d)
BUILD_VERSION ?= $(shell git describe --always)
BUILD_TAG ?= $(BUILD_DATE)-$(BUILD_VERSION)
BUILD_OS ?= $(shell uname -s)
BUILD_ARCH ?= $(shell scripts/dist/arch.sh)
JS_BUILD_PATH ?= $(shell realpath "./assets/static/build")

# Installation Parameters
INSTALL_PATH ?= $(BUILD_PATH)/photoprism-$(BUILD_TAG)-$(shell echo $(BUILD_OS) | tr '[:upper:]' '[:lower:]')-$(BUILD_ARCH)
DESTDIR ?= $(INSTALL_PATH)
DESTUID ?= 1000
DESTGID ?= 1000
INSTALL_USER ?= $(DESTUID):$(DESTGID)
INSTALL_MODE ?= u+rwX,a+rX
INSTALL_MODE_BIN ?= 755

UID := $(shell id -u)
HASRICHGO := $(shell which richgo)

ifdef HASRICHGO
    GOTEST=richgo test
else
    GOTEST=go test
endif

all: dep build
dep: dep-tensorflow dep-npm dep-js dep-go
build: build-js
test: test-js test-go
test-go: reset-testdb run-test-go
test-pkg: reset-testdb run-test-pkg
test-api: reset-testdb run-test-api
test-short: reset-testdb run-test-short
acceptance-private-run-chromium: acceptance-private-restart acceptance-private acceptance-private-stop
acceptance-public-run-chromium: acceptance-restart acceptance acceptance-stop
acceptance-private-run-firefox: acceptance-private-restart acceptance-private-firefox acceptance-private-stop
acceptance-public-run-firefox: acceptance-restart acceptance-firefox acceptance-stop
acceptance-run-chromium: acceptance-private-restart acceptance-private acceptance-private-stop acceptance-restart acceptance acceptance-stop
acceptance-run-firefox: acceptance-private-restart acceptance-private-firefox acceptance-private-stop acceptance-restart acceptance-firefox acceptance-stop
test-all: test acceptance-run-chromium
fmt: fmt-js fmt-go
clean-local: clean-local-config clean-local-cache
upgrade: dep-upgrade-js dep-upgrade
devtools: install-go dep-npm
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
install:
	$(info Installing in "$(DESTDIR)"...)
	[ ! -d "$(DESTDIR)" ] || rm -rf --preserve-root $(DESTDIR)
	mkdir --mode=$(INSTALL_MODE) -p $(DESTDIR)
	env TMPDIR="$(BUILD_PATH)" ./scripts/dist/install-tensorflow.sh $(DESTDIR)
	rm -rf --preserve-root $(DESTDIR)/include
	(cd $(DESTDIR) && mkdir -p bin scripts lib assets config config/examples)
	scripts/build.sh prod $(DESTDIR)/bin/$(BINARY_NAME)
	[ -f "$(GOBIN)/gosu" ] || go install github.com/tianon/gosu@latest
	cp $(GOBIN)/gosu $(DESTDIR)/bin/gosu
	[ -f "$(GOBIN)/exif-read-tool" ] || go install github.com/dsoprea/go-exif/v3/command/exif-read-tool@latest
	cp $(GOBIN)/exif-read-tool $(DESTDIR)/bin/exif-read-tool
	rsync -r -l --safe-links --exclude-from=assets/.buildignore --chmod=a+r,u+rw ./assets/ $(DESTDIR)/assets
	rsync -r -l --safe-links --exclude-from=scripts/dist/.buildignore --chmod=a+rx,u+rwx ./scripts/dist/ $(DESTDIR)/scripts
	mv $(DESTDIR)/scripts/heif-convert.sh $(DESTDIR)/bin/heif-convert
	cp internal/config/testdata/*.yml $(DESTDIR)/config/examples
	chown -R $(INSTALL_USER) $(DESTDIR)
	chmod -R $(INSTALL_MODE) $(DESTDIR)
	chmod -R $(INSTALL_MODE_BIN) $(DESTDIR)/bin $(DESTDIR)/lib $(DESTDIR)/scripts/*.sh
	echo "PhotoPrism $(BUILD_TAG) has been successfully installed in \"$(DESTDIR)\".\nEnjoy!"
install-go:
	sudo scripts/dist/install-go.sh
	go build -v ./...
install-tensorflow:
	sudo scripts/dist/install-tensorflow.sh
install-darktable:
	sudo scripts/dist/install-darktable.sh
acceptance-restart:
	cp -f storage/acceptance/backup.db storage/acceptance/index.db
	cp -f storage/acceptance/config/settingsBackup.yml storage/acceptance/config/settings.yml
	rm -rf storage/acceptance/sidecar/2020
	rm -rf storage/acceptance/sidecar/2011
	rm -rf storage/acceptance/originals/2010
	rm -rf storage/acceptance/originals/2020
	rm -rf storage/acceptance/originals/2011
	rm -rf storage/acceptance/originals/2013
	rm -rf storage/acceptance/originals/2017
	go run cmd/photoprism/photoprism.go --public --upload-nsfw=false --database-driver sqlite --database-dsn ./storage/acceptance/index.db --import-path ./storage/acceptance/import --http-port=2343 --config-path ./storage/acceptance/config --originals-path ./storage/acceptance/originals --storage-path ./storage/acceptance --test --backup-path ./storage/acceptance/backup --disable-backups start -d
acceptance-stop:
	go run cmd/photoprism/photoprism.go --public --upload-nsfw=false --database-driver sqlite --database-dsn ./storage/acceptance/index.db --import-path ./storage/acceptance/import --http-port=2343 --config-path ./storage/acceptance/config --originals-path ./storage/acceptance/originals --storage-path ./storage/acceptance --test --backup-path ./storage/acceptance/backup --disable-backups stop
acceptance-private-restart:
	cp -f storage/acceptance/backup.db storage/acceptance/index.db
	cp -f storage/acceptance/config/settingsBackup.yml storage/acceptance/config/settings.yml
	go run cmd/photoprism/photoprism.go --public=false --upload-nsfw=false --database-driver sqlite --database-dsn ./storage/acceptance/index.db --import-path ./storage/acceptance/import --http-port=2343 --config-path ./storage/acceptance/config --originals-path ./storage/acceptance/originals --storage-path ./storage/acceptance --test --backup-path ./storage/acceptance/backup --disable-backups start -d
acceptance-private-stop:
	go run cmd/photoprism/photoprism.go --public=false --upload-nsfw=false --database-driver sqlite --database-dsn ./storage/acceptance/index.db --import-path ./storage/acceptance/import --http-port=2343 --config-path ./storage/acceptance/config --originals-path ./storage/acceptance/originals --storage-path ./storage/acceptance --test --backup-path ./storage/acceptance/backup --disable-backups stop
start:
	go run cmd/photoprism/photoprism.go start -d
stop:
	go run cmd/photoprism/photoprism.go stop
terminal:
	docker-compose exec -u $(UID) photoprism bash
root-terminal:
	docker-compose exec -u root photoprism bash
migrate:
	go run cmd/photoprism/photoprism.go migrate
generate:
	go generate ./pkg/... ./internal/...
	go fmt ./pkg/... ./internal/...
	# Revert unnecessary file change?
	POT_UNCHANGED='1 file changed, 1 insertion(+), 1 deletion(-)'
	@if [ ${$(shell git diff --shortstat assets/locales/messages.pot):1:45} == $(POT_UNCHANGED) ]; then\
		git checkout -- assets/locales/messages.pot;\
	fi
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
	(cd frontend &&	npm install --silent --legacy-peer-deps)
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
	(cd frontend &&	env NODE_ENV=development BABEL_ENV=test npm run test)
acceptance:
	$(info Running JS acceptance tests in Chrome...)
	(cd frontend &&	npm run acceptance && cd ..)
acceptance-firefox:
	$(info Running JS acceptance tests in Firefox...)
	(cd frontend &&	npm run acceptance-firefox && cd ..)
acceptance-private:
	$(info Running JS acceptance-private tests in Chrome...)
	(cd frontend &&	npm run acceptance-private && cd ..)
acceptance-private-firefox:
	$(info Running JS acceptance-private tests in Firefox...)
	(cd frontend &&	npm run acceptance-private-firefox && cd ..)
reset-mariadb:
	$(info Resetting photoprism database...)
	mysql < scripts/sql/reset-mariadb.sql
reset-testdb:
	$(info Removing test database files...)
	find ./internal -type f -name ".test.*" -delete
run-test-short:
	$(info Running short Go unit tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -short -timeout 5m ./pkg/... ./internal/...
run-test-go:
	$(info Running all Go unit tests...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
run-test-pkg:
	$(info Running all Go unit tests in "/pkg"...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./pkg/...
run-test-api:
	$(info Running all API unit tests...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./internal/api/...
test-parallel:
	$(info Running all Go unit tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -tags slow -timeout 20m ./pkg/... ./internal/...
test-verbose:
	$(info Running all Go unit tests in verbose mode...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m -v ./pkg/... ./internal/...
test-race:
	$(info Running all Go unit tests with race detection in verbose mode...)
	$(GOTEST) -tags slow -race -timeout 60m -v ./pkg/... ./internal/...
test-codecov:
	$(info Running all Go unit tests with code coverage report for codecov...)
	go test -parallel 1 -count 1 -cpu 1 -failfast -tags slow -timeout 30m -coverprofile coverage.txt -covermode atomic ./pkg/... ./internal/...
	scripts/codecov.sh -t $(CODECOV_TOKEN)
test-coverage:
	$(info Running all Go unit tests with code coverage report...)
	go test -parallel 1 -count 1 -cpu 1 -failfast -tags slow -timeout 30m -coverprofile coverage.txt -covermode atomic ./pkg/... ./internal/...
	go tool cover -html=coverage.txt -o coverage.html
docker-develop-all: docker-develop-bullseye docker-develop-armv7 docker-develop-buster docker-develop-impish
docker-develop: docker-develop-bullseye
docker-develop-bullseye:
	docker pull --platform=amd64 golang:bullseye
	docker pull --platform=arm64 golang:bullseye
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bullseye /bullseye "-t photoprism/develop:latest"
docker-develop-armv7:
	docker pull --platform=arm golang:bullseye
	scripts/docker/buildx.sh develop linux/arm armv7 /armv7
docker-develop-buster:
	docker pull --platform=amd64 golang:buster
	docker pull --platform=arm64 golang:buster
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 buster /buster
docker-develop-impish:
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 impish /impish
docker-preview-all: docker-preview-bullseye docker-preview-buster docker-preview-impish
docker-preview: docker-preview-bullseye
docker-preview-bullseye:
	docker pull --platform=amd64 photoprism/develop:bullseye
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=amd64 debian:bullseye-slim
	docker pull --platform=arm64 debian:bullseye-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview /bullseye
docker-preview-armv7:
	docker pull --platform=arm photoprism/develop:armv7
	docker pull --platform=arm debian:bullseye-slim
	scripts/docker/buildx.sh photoprism linux/arm preview-armv7 /armv7
docker-preview-arm64:
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=arm64 debian:bullseye-slim
	scripts/docker/buildx.sh photoprism linux/arm64 preview-arm64 /bullseye
docker-preview-buster:
	docker pull --platform=amd64 photoprism/develop:buster
	docker pull --platform=arm64 photoprism/develop:buster
	docker pull --platform=amd64 debian:buster-slim
	docker pull --platform=arm64 debian:buster-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-buster /buster
docker-preview-impish:
	docker pull --platform=amd64 photoprism/develop:latest
	docker pull --platform=arm64 photoprism/develop:latest
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 preview-impish /impish
docker-release-all: docker-release-bullseye docker-release-buster docker-release-impish
docker-release: docker-release-bullseye
docker-release-bullseye:
	docker pull --platform=amd64 photoprism/develop:bullseye
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=amd64 debian:bullseye-slim
	docker pull --platform=arm64 debian:bullseye-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 bullseye /bullseye "-t photoprism/photoprism:latest"
docker-release-armv7:
	docker pull --platform=arm photoprism/develop:armv7
	docker pull --platform=arm debian:bullseye-slim
	scripts/docker/buildx.sh photoprism linux/arm armv7 /armv7
docker-release-arm64:
	docker pull --platform=arm64 photoprism/develop:bullseye
	docker pull --platform=arm64 debian:bullseye-slim
	scripts/docker/buildx.sh photoprism linux/arm64 arm64 /bullseye
docker-release-buster:
	docker pull --platform=amd64 photoprism/develop:buster
	docker pull --platform=arm64 photoprism/develop:buster
	docker pull --platform=amd64 debian:buster-slim
	docker pull --platform=arm64 debian:buster-slim
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 buster /buster
docker-release-impish:
	docker pull --platform=amd64 photoprism/develop:impish
	docker pull --platform=arm64 photoprism/develop:impish
	docker pull --platform=amd64 ubuntu:impish
	docker pull --platform=arm64 ubuntu:impish
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 impish /impish
docker-local:
	scripts/docker/build.sh photoprism
docker-pull:
	docker pull photoprism/photoprism:preview photoprism/photoprism:latest
docker-ddns:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh ddns linux/amd64,linux/arm64 $(BUILD_DATE)
docker-goproxy:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh goproxy linux/amd64,linux/arm64 $(BUILD_DATE)
docker-demo:
	docker pull photoprism/photoprism:preview
	scripts/docker/build.sh demo $(BUILD_DATE)
	scripts/docker/push.sh demo $(BUILD_DATE)
docker-demo-local:
	scripts/docker/build.sh photoprism
	scripts/docker/build.sh demo $(BUILD_DATE)
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
	(cd ./docker/examples/cloud && packer build digitalocean.json)
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
	go mod tidy

.PHONY: all build dev dep-npm dep dep-go dep-js dep-list dep-tensorflow dep-upgrade dep-upgrade-js test test-js test-go \
    install generate fmt fmt-go fmt-js upgrade start stop terminal root-terminal packer-digitalocean acceptance clean tidy \
    install-go install-darktable install-tensorflow devtools tar.gz;