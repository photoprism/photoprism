export GO111MODULE=on
GOCMD=go
GOINSTALL=$(GOCMD) install
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=photoprism

all: deps test build
install:
	$(GOINSTALL) cmd/photoprism/photoprism.go
build:
	$(GOBUILD) cmd/photoprism/photoprism.go -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
image:
	docker build . --tag photoprism/photoprism
	docker push photoprism/photoprism
deps:
	$(GOBUILD) -v ./...