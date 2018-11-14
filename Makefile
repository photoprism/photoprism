export GO111MODULE=on
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOIMPORTS=goimports
BINARY_NAME=photoprism
DOCKER_TAG=`date -u +%Y%m%d`

all: download dep js build
install: install-bin install-assets install-config
install-bin:
	scripts/build.sh prod /usr/local/bin/$(BINARY_NAME)
install-assets:
	mkdir -p /srv/photoprism/photos
	mkdir -p /srv/photoprism/cache
	mkdir -p /srv/photoprism/database
	cp -r assets/server /srv/photoprism
	cp -r assets/tensorflow /srv/photoprism
	find /srv/photoprism -name '.*' -type f -delete
install-config:
	mkdir -p /etc/photoprism
	test -e /etc/photoprism/photoprism.yml || cp -n configs/photoprism.yml /etc/photoprism/photoprism.yml
build:
	scripts/build.sh debug $(BINARY_NAME)
js:
	(cd frontend &&	yarn install --frozen-lockfile --prod)
	(cd frontend &&	env NODE_ENV=production npm run build)
start:
	$(GORUN) cmd/photoprism/photoprism.go start
migrate:
	$(GORUN) cmd/photoprism/photoprism.go migrate
test:
	$(GOTEST) -tags=slow -timeout 20m -v ./internal/...
test-fast:
	$(GOTEST) -timeout 20m -v ./internal/...
test-race:
	$(GOTEST) -tags=slow -race -timeout 60m -v ./internal/...
test-codecov:
	$(GOTEST) -tags=slow -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
	scripts/codecov.sh
test-coverage:
	$(GOTEST) -tags=slow -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
	$(GOTOOL) cover -html=coverage.txt -o coverage.html
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
download:
	scripts/download-inception.sh
deploy-photoprism:
	scripts/docker-build.sh photoprism $(DOCKER_TAG)
	scripts/docker-push.sh photoprism $(DOCKER_TAG)
deploy-development:
	scripts/docker-build.sh development $(DOCKER_TAG)
	scripts/docker-push.sh development $(DOCKER_TAG)
deploy-tensorflow:
	scripts/docker-build.sh tensorflow $(DOCKER_TAG)
	scripts/docker-push.sh tensorflow $(DOCKER_TAG)
deploy-darktable:
	DARKTABLE_VERSION="$(awk '$2 == "DARKTABLE_VERSION" { print $3; exit }' docker/darktable/Dockerfile)"
	scripts/docker-build.sh darktable $(DARKTABLE_VERSION)
	scripts/docker-push.sh darktable $(DARKTABLE_VERSION)
fmt:
	$(GOIMPORTS) -w internal cmd
	$(GOFMT) ./internal/... ./cmd/...
dep:
	$(GOBUILD) -v ./...
	$(GOMOD) tidy
upgrade:
	$(GOGET) -u