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

all: dep js build
install:
	$(GOINSTALL) cmd/photoprism/photoprism.go
build:
	$(GOBUILD) cmd/photoprism/photoprism.go
js:
	(cd frontend &&	yarn install)
	(cd frontend &&	npm run build)
start:
	$(GORUN) cmd/photoprism/photoprism.go start
migrate-db:
	$(GORUN) cmd/photoprism/photoprism.go migrate-db
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
image:
	docker build . --tag photoprism/photoprism
	docker push photoprism/photoprism
fmt:
	$(GOFMT) ./...
dep:
	$(GOBUILD) -v ./...
upgrade:
	$(GOGET) -u