export GO111MODULE=on
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOIMPORTS=goimports
BINARY_NAME=photoprism

all: tensorflow-model dep js build
install: install-bin install-assets install-config
install-bin:
	$(GOINSTALL) cmd/photoprism/photoprism.go
install-assets:
	mkdir -p /var/photoprism
	mkdir -p /var/photoprism/photos
	mkdir -p /var/photoprism/thumbnails
	cp -r assets/favicons /var/photoprism
	cp -r assets/public /var/photoprism
	cp -r assets/templates /var/photoprism
	cp -r assets/tensorflow /var/photoprism
install-config:
	mkdir -p /etc/photoprism
	test -e /etc/photoprism/photoprism.yml || cp -n configs/photoprism.yml /etc/photoprism/photoprism.yml
build:
	$(GOBUILD) cmd/photoprism/photoprism.go
js:
	(cd frontend &&	yarn install --prod)
	(cd frontend &&	env NODE_ENV=production npm run build)
start:
	$(GORUN) cmd/photoprism/photoprism.go start
migrate:
	$(GORUN) cmd/photoprism/photoprism.go migrate
test:
	$(GOTEST) -v ./internal/...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
tensorflow-model:
	scripts/download-tf-model.sh
docker-push:
	scripts/docker-push.sh
fmt:
	$(GOIMPORTS) -w internal cmd
	$(GOFMT) ./internal/... ./cmd/...
dep:
	$(GOBUILD) -v ./...
	$(GOMOD) tidy
upgrade:
	$(GOGET) -u