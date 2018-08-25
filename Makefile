# Go parameters
GOCMD=go
GOLINT=golint
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BUILD_TARGET_MAIN=main
#BUILD TARGET

all: build

build: fmt  #TODO: change it to all relevant tools

.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

.PHONY: test
test:
	$(GOTEST) -short ./...

.PHONY: lint
lint: 
	go list ./... | grep -v "^.*exchangeimpl/.*/.*" | xargs $(GOLINT)

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f ./bin/*
