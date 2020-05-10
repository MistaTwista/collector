#!make
include .env
export $(shell sed 's/=.*//' .env)

VERSION=$(shell git describe --tags)
BUILD=$(shell git rev-parse --short HEAD)
PROJECTNAME=$(shell basename "$(PWD)")

REGISTRY?=gcr.io/images

PKGS := $(shell go list ./... | grep -v /vendor)

# Go related variables.
GOBASE=$(shell pwd)
GOPATH=$(GOBASE)/vendor:$(GOBASE)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Redirect error output to a file, so we can show it in development mode.
STDERR=tmp/.$(PROJECTNAME)-stderr.txt

# PID file will store the server process id when it's running on development mode
PID=tmp/.$(PROJECTNAME)-process.pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: install
install: go-get ## Install dependencies

.PHONY: install-linter
install-linter:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: install-linter ## Run linter
	@golangci-lint run

.PHONY: clean
clean: ## GO Clean
	go clean

.PHONY: build
build: build-play ## Build a binary
	GOOS=linux GOARCH=386 go build -o bin/${PROJECTNAME} main.go

.PHONY: build-play
build-play:
	${MAKE} -C playground build

.PHONY: replay
replay: build-play ## Replay
	./bin/json-play

.PHONY: run
run: ## Run current version
	./bin/${PROJECTNAME}

.PHONY: rerun
rerun: build run ## Build and Run

.PHONY: test
test: lint ## Run tests
	go test -v $(PKGS)

.PHONY: docker-build
docker-build: build ## Build docker container
	docker build -t ${PROJECTNAME} .
	docker tag ${PROJECTNAME} ${PROJECTNAME}:${BUILD}

.PHONY: docker-push
docker-push: check-environment docker-build ## Push container to repo
	docker push ${REGISTRY}/${ENV}/${PROJECTNAME}:${BUILD}

.PHONY: help
.DEFAULT_GOAL := help
help: ## Print this help message
	@echo "Usage:"
	@echo "------"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

check-environment:
ifndef ENV
	$(error ENV not set, allowed values - `staging` or `production`)
endif

.PHONY: printvars
printvars:
	$(foreach v, $(filter-out .VARIABLES,$(.VARIABLES)), $(info $(v) = $($(v))))

.PHONY: echo
echo: ## Show env
	env

tree: ## Show tree
	tree -I vendor

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  >  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

start:
	@bash -c "trap 'make stop' EXIT; $(MAKE) clean compile start-server watch run='make clean compile start-server'"

stop: stop-server

start-server: stop-server
	@echo "  >  $(PROJECTNAME) is available at $(ADDR)"
	@-$(GOBIN)/$(PROJECTNAME) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"

stop-server:
	@-touch $(PID)
	@-kill `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

watch:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) yolo -i . -e vendor -e bin -c "$(run)"

restart-server: stop-server start-server
