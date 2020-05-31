export GO111MODULE=on
GOIMPORTS=goimports
BINARY_NAME=photoprism
DOCKER_TAG=`date -u +%Y%m%d`

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
test-go: reset-test-db run-test-go
test-short: reset-test-db run-test-short
acceptance-all: start acceptance acceptance-firefox stop
test-all: test acceptance-all
fmt: fmt-js fmt-go
upgrade: dep-upgrade-js dep-upgrade
clean-local: clean-local-config clean-local-share clean-local-cache
clean-install: clean-local dep build-js install-bin install-assets
start:
	go run cmd/photoprism/photoprism.go start -d
stop:
	go run cmd/photoprism/photoprism.go stop
terminal:
	docker-compose exec photoprism bash
migrate:
	go run cmd/photoprism/photoprism.go migrate
generate:
	go generate ./pkg/... ./internal/...
	go fmt ./pkg/... ./internal/...
install-bin:
	scripts/build.sh prod ~/.local/bin/$(BINARY_NAME)
install-assets:
	$(info Installing assets)
	mkdir -p ~/.photoprism/storage/settings
	mkdir -p ~/.photoprism/storage/cache
	mkdir -p ~/.photoprism/storage
	mkdir -p ~/.photoprism/assets
	mkdir -p ~/Pictures/Originals
	mkdir -p ~/Pictures/Import
	cp -r assets/static assets/templates assets/nasnet assets/nsfw ~/.photoprism/assets
	find ~/.photoprism/assets -name '.*' -type f -delete
clean-local-assets:
	rm -rf ~/.photoprism/assets/*
clean-local-cache:
	rm -rf ~/.photoprism/storage/cache/*
clean-local-config:
	rm -f ~/.photoprism/storage/settings/*
dep-js:
	(cd frontend &&	npm install --silent)
dep-go:
	go build -v ./...
dep-upgrade:
	go get -u -t ./...
dep-upgrade-js:
	(cd frontend &&	npm --depth 3 update)
dep-tensorflow:
	scripts/download-nasnet.sh
	scripts/download-nsfw.sh
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
	(cd frontend &&	npm run acceptance)
acceptance-firefox:
	$(info Running JS acceptance tests in Firefox...)
	(cd frontend &&	npm run acceptance-firefox)
reset-test-db:
	$(info Purging test databases...)
	mysql < scripts/reset-test-db.sql
	find ./internal -type f -name '.test.*' -delete
run-test-short:
	$(info Running short Go unit tests in parallel mode...)
	$(GOTEST) -parallel 2 -count 1 -cpu 2 -short -timeout 5m ./pkg/... ./internal/...
run-test-go:
	$(info Running all Go unit tests...)
	$(GOTEST) -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...
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
	scripts/codecov.sh
test-coverage:
	$(info Running all Go unit tests with code coverage report...)
	go test -parallel 1 -count 1 -cpu 1 -failfast -tags slow -timeout 30m -coverprofile coverage.txt -covermode atomic ./pkg/... ./internal/...
	go tool cover -html=coverage.txt -o coverage.html
clean:
	rm -f $(BINARY_NAME)
	rm -f *.log
	rm -rf node_modules
	rm -rf storage/testdata
	rm -rf storage/backups
	rm -rf storage/cache
	rm -rf frontend/node_modules
docker-development:
	scripts/docker-build.sh development $(DOCKER_TAG)
	scripts/docker-push.sh development $(DOCKER_TAG)
docker-photoprism:
	scripts/docker-build.sh photoprism $(DOCKER_TAG)
	scripts/docker-push.sh photoprism $(DOCKER_TAG)
docker-photoprism-arm64:
	scripts/docker-build.sh photoprism-arm64 $(DOCKER_TAG)
	scripts/docker-push.sh photoprism-arm64 $(DOCKER_TAG)
docker-demo:
	scripts/docker-build.sh demo $(DOCKER_TAG)
	scripts/docker-push.sh demo $(DOCKER_TAG)
docker-webdav:
	scripts/docker-build.sh webdav $(DOCKER_TAG)
	scripts/docker-push.sh webdav $(DOCKER_TAG)
lint-js:
	(cd frontend &&	npm run lint)
fmt-js:
	(cd frontend &&	npm run fmt)
fmt-imports:
	goimports -w pkg internal cmd
fmt-go:
	go fmt ./pkg/... ./internal/... ./cmd/...
tidy:
	go mod tidy
