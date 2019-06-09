export GO111MODULE=on
GOIMPORTS=goimports
BINARY_NAME=photoprism
DOCKER_TAG=`date -u +%Y%m%d`
TIDB_VERSION=2.1.11
DARKTABLE_VERSION="$(awk '$2 == "DARKTABLE_VERSION" { print $3; exit }' docker/darktable/Dockerfile)"

HASRICHGO := $(shell which richgo)
ifdef HASRICHGO
    GOTEST=richgo test
else
    GOTEST=go test
endif

all: dep build
dep: dep-tensorflow dep-js dep-go
build: build-js build-go
install: install-bin install-assets
test: test-js test-go
fmt: fmt-js fmt-go
upgrade: upgrade-js upgrade-go
start:
	go run cmd/photoprism/photoprism.go start
migrate:
	go run cmd/photoprism/photoprism.go migrate
install-bin:
	$(info Building prodution binary...)
	scripts/build.sh prod /usr/local/bin/$(BINARY_NAME)
install-assets:
	$(info Installing assets in /srv/photoprism...)
	mkdir -p /srv/photoprism/config
	mkdir -p /srv/photoprism/photos
	mkdir -p /srv/photoprism/cache
	mkdir -p /srv/photoprism/resources/database
	cp -r assets/resources/static assets/resources/templates assets/resources/nasnet /srv/photoprism/resources
	rsync -a -v --ignore-existing assets/config/*.yml /srv/photoprism/config
	find /srv/photoprism -name '.*' -type f -delete
dep-js:
	(cd frontend &&	npm install)
dep-go:
	go build -v ./...
dep-tensorflow:
	scripts/download-nasnet.sh
zip-nasnet:
	(cd assets/resources && zip -r nasnet.zip nasnet -x "*/.*" -x "*/version.txt")
build-js:
	(cd frontend &&	env NODE_ENV=production npm run build)
build-go:
	rm -f $(BINARY_NAME)
	scripts/build.sh debug $(BINARY_NAME)
watch-js:
	(cd frontend &&	env NODE_ENV=development npm run watch)
test-js:
	$(info Running JS unit tests...)
	(cd frontend &&	env NODE_ENV=development npm run test)
test-chromium:
	$(info Running JS acceptance tests in Chrome...)
	(cd frontend &&	npm run test-chromium)
test-firefox:
	$(info Running JS acceptance tests in Firefox...)
	(cd frontend &&	npm run test-firefox)
test-go:
	$(info Running all Go unit tests...)
	$(GOTEST) -tags=slow -timeout 20m ./internal/...
test-verbose:
	$(info Running all Go unit tests in verbose mode...)
	$(GOTEST) -tags=slow -timeout 20m -v ./internal/...
test-short:
	$(info Running short Go unit tests in verbose mode...)
	$(GOTEST) -short -timeout 5m -v ./internal/...
test-race:
	$(info Running all Go unit tests with race detection in verbose mode...)
	$(GOTEST) -tags=slow -race -timeout 60m -v ./internal/...
test-codecov:
	$(info Running all Go unit tests with code coverage report for codecov...)
	go test -tags=slow -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
	scripts/codecov.sh
test-coverage:
	$(info Running all Go unit tests with code coverage report...)
	go test -tags=slow -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
	go tool cover -html=coverage.txt -o coverage.html
clean:
	rm -f $(BINARY_NAME)
	rm -f *.log
	rm -rf node_modules
	rm -rf assets/testdata
	rm -rf assets/backups
	rm -rf frontend/node_modules
docker-development:
	scripts/docker-build.sh development $(DOCKER_TAG)
	scripts/docker-push.sh development $(DOCKER_TAG)
docker-photoprism:
	scripts/docker-build.sh photoprism $(DOCKER_TAG)
	scripts/docker-push.sh photoprism $(DOCKER_TAG)
docker-demo:
	scripts/docker-build.sh demo $(DOCKER_TAG)
	scripts/docker-push.sh demo $(DOCKER_TAG)
docker-tensorflow:
	scripts/docker-build.sh tensorflow $(DOCKER_TAG)
	scripts/docker-push.sh tensorflow $(DOCKER_TAG)
docker-darktable:
	scripts/docker-build.sh darktable $(DARKTABLE_VERSION)
	scripts/docker-push.sh darktable $(DARKTABLE_VERSION)
docker-tidb:
	scripts/docker-build.sh tidb $(TIDB_VERSION)
	scripts/docker-push.sh tidb $(TIDB_VERSION)
lint-js:
	(cd frontend &&	npm run lint)
fmt-js:
	(cd frontend &&	npm run fmt)
fmt-go:
	goimports -w internal cmd
	go fmt ./internal/... ./cmd/...
tidy:
	go mod tidy
upgrade-js:
	(cd frontend &&	npm update --depth 1)
upgrade-go:
	go mod tidy
	go get -u
