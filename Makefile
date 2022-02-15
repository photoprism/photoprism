.PHONY: all build dev npm dep dep-go dep-js dep-list dep-tensorflow dep-upgrade dep-upgrade-js \
		test test-js test-go install generate fmt fmt-go fmt-js upgrade start stop \
		terminal root-terminal packer-digitalocean acceptance clean tidy;
.SILENT: ;               # no need for @
.ONESHELL: ;             # recipes execute in same shell
.NOTPARALLEL: ;          # wait for target to finish
.EXPORT_ALL_VARIABLES: ; # send all vars to shell

export GO111MODULE=on
export GODEBUG=http2client=0

GOIMPORTS=goimports
BINARY_NAME=photoprism

DOCKER_TAG := $(shell date -u +%Y%m%d)
UID := $(shell id -u)
HASRICHGO := $(shell which richgo)

ifdef HASRICHGO
    GOTEST=richgo test
else
    GOTEST=go test
endif

all: dep build
dep: dep-tensorflow dep-js dep-go
build: generate build-js build-go
install: install-bin install-assets
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
upgrade: dep-upgrade-js dep-upgrade
clean-local: clean-local-config clean-local-cache
clean-install: clean-local dep build-js install-bin install-assets
dev: npm install-go-amd64
npm:
	$(info Upgrading NPM version...)
	sudo npm update -g npm
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
install-go-amd64:
	$(info Installing Go (AMD64)...)
	sudo docker/scripts/install-go.sh amd64
	go build -v ./...
install-bin:
	scripts/build.sh prod ~/.local/bin/$(BINARY_NAME)
install-assets:
	$(info Installing assets)
	mkdir -p ~/.photoprism/storage/config
	mkdir -p ~/.photoprism/storage/cache
	mkdir -p ~/.photoprism/storage
	mkdir -p ~/.photoprism/assets
	mkdir -p ~/Pictures/Originals
	mkdir -p ~/Pictures/Import
	cp -r assets/locales assets/facenet assets/nasnet assets/nsfw assets/profiles assets/static assets/templates ~/.photoprism/assets
	find ~/.photoprism/assets -name '.*' -type f -delete
clean-local-assets:
	rm -rf ~/.photoprism/assets/*
clean-local-cache:
	rm -rf ~/.photoprism/storage/cache/*
clean-local-config:
	rm -f ~/.photoprism/storage/config/*
dep-list:
	go list -u -m -json all | go-mod-outdated -direct
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
build-go:
	rm -f $(BINARY_NAME)
	scripts/build.sh debug $(BINARY_NAME)
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
	find ./internal -type f -name '.test.*' -delete
run-test-short:
	$(info Running short Go unit tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -short -timeout 5m ./pkg/... ./internal/...
run-test-go:
	$(info Running all Go unit tests...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
run-test-pkg:
	$(info Running all Go unit tests in '/pkg'...)
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
clean:
	rm -f $(BINARY_NAME)
	rm -f *.log
	rm -rf node_modules
	rm -rf storage/testdata
	rm -rf storage/backup
	rm -rf storage/cache
	rm -rf frontend/node_modules
docker-develop:
	docker pull --platform=amd64 ubuntu:21.10
	docker pull --platform=arm64 ubuntu:21.10
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 $(DOCKER_TAG)
docker-develop-buster:
	docker pull --platform=amd64 golang:buster
	docker pull --platform=arm64 golang:buster
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 buster /buster
docker-develop-bullseye:
	docker pull --platform=amd64 golang:bullseye
	docker pull --platform=arm64 golang:bullseye
	scripts/docker/buildx-multi.sh develop linux/amd64,linux/arm64 bullseye /bullseye
docker-preview:
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64
docker-release:
	scripts/docker/buildx-multi.sh photoprism linux/amd64,linux/arm64 $(DOCKER_TAG)
docker-arm64-preview:
	scripts/docker/buildx.sh photoprism linux/arm64 arm64-preview
docker-arm64-release:
	scripts/docker/buildx.sh photoprism linux/arm64 arm64
docker-armv7-develop:
	docker pull --platform=arm ubuntu:21.10
	scripts/docker/buildx.sh develop linux/arm armv7 /armv7
docker-armv7-preview:
	docker pull --platform=arm photoprism/develop:armv7
	scripts/docker/buildx.sh photoprism linux/arm armv7-preview /armv7
docker-armv7-release:
	docker pull --platform=arm photoprism/develop:armv7
	scripts/docker/buildx.sh photoprism linux/arm armv7 /armv7
docker-local:
	scripts/docker/build.sh photoprism
docker-pull:
	docker pull photoprism/photoprism:preview photoprism/photoprism:latest
docker-ddns:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh ddns linux/amd64,linux/arm64 $(DOCKER_TAG)
docker-goproxy:
	docker pull golang:alpine
	scripts/docker/buildx-multi.sh goproxy linux/amd64,linux/arm64 $(DOCKER_TAG)
docker-demo:
	scripts/docker/build.sh demo $(DOCKER_TAG)
	scripts/docker/push.sh demo $(DOCKER_TAG)
docker-demo-local:
	scripts/docker/build.sh photoprism
	scripts/docker/build.sh demo $(DOCKER_TAG)
	scripts/docker/push.sh demo $(DOCKER_TAG)
docker-dummy-webdav:
	docker pull --platform=amd64 golang:1
	docker pull --platform=arm64 golang:1
	scripts/docker/buildx-multi.sh dummy-webdav linux/amd64,linux/arm64 $(DOCKER_TAG)
docker-dummy-oidc:
	docker pull --platform=amd64 golang:1
	docker pull --platform=arm64 golang:1
	scripts/docker/buildx-multi.sh dummy-oidc linux/amd64,linux/arm64 $(DOCKER_TAG)
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
