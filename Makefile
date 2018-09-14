export GO111MODULE=on
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
BINARY_NAME=photoprism

all: tensorflow-model dep js build
install: install-bin install-assets install-config
install-bin:
	$(GOINSTALL) cmd/photoprism/photoprism.go
install-assets:
	cp -r assets /var/photoprism
install-config:
	mkdir -p /etc/photoprism
	cp configs/photoprism.prod.yml /etc/photoprism/photoprism.yml
build:
	$(GOBUILD) cmd/photoprism/photoprism.go
js:
	(cd frontend &&	yarn install --prod)
	(cd frontend &&	env NODE_ENV=production npm run build)
start:
	$(GORUN) cmd/photoprism/photoprism.go start
migrate-db:
	$(GORUN) cmd/photoprism/photoprism.go migrate-db
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
tensorflow-model:
	scripts/download-tf-model.sh
image:
	docker build . --tag photoprism/photoprism
	docker push photoprism/photoprism
fmt:
	$(GOFMT) ./...
dep:
	$(GOBUILD) -v ./...
upgrade:
	$(GOGET) -u