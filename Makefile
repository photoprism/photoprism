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
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
install-assets:
	mkdir -p /srv/photoprism
	mkdir -p /srv/photoprism/photos
	mkdir -p /srv/photoprism/thumbnails
	cp -r assets/favicons /srv/photoprism
	cp -r assets/public /srv/photoprism
	cp -r assets/templates /srv/photoprism
	cp -r assets/tensorflow /srv/photoprism
install-config:
	mkdir -p /etc/photoprism
	test -e /etc/photoprism/photoprism.yml || cp -n configs/photoprism.yml /etc/photoprism/photoprism.yml
build:
	scripts/build.sh
js:
	(cd frontend &&	yarn install --prod)
	(cd frontend &&	env NODE_ENV=production npm run build)
start:
	$(GORUN) cmd/photoprism/photoprism.go start
migrate:
	$(GORUN) cmd/photoprism/photoprism.go migrate
test:
	$(GOTEST) -timeout 20m -v ./internal/...
test-race:
	$(GOTEST) -race -timeout 60m -v ./internal/...
test-codecov:
	$(GOTEST) -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
	scripts/codecov.sh
test-coverage:
	$(GOTEST) -timeout 30m -coverprofile=coverage.txt -covermode=atomic -v ./internal/...
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
fmt:
	$(GOIMPORTS) -w internal cmd
	$(GOFMT) ./internal/... ./cmd/...
dep:
	$(GOBUILD) -v ./...
	$(GOMOD) tidy
upgrade:
	$(GOGET) -u