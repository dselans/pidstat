BIN             = pidstat
OUTPUT_DIR      = build
RELEASE_TIME   := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
RELEASE_VER    := $(shell git rev-parse --short HEAD)
LDFLAGS        := "-s -w -X main.version=$(RELEASE_VER)-$(RELEASE_TIME)"
COVERMODE       = atomic

TEST_PACKAGES      := $(shell go list ./... | grep -v vendor | grep -v fakes | grep -v ftest)

.PHONY: help docs
.DEFAULT_GOAL := help

run: ## Run application (without building)
	go run *.go

test: ## Run tests
	go test $(TEST_PACKAGES)

test/cover: ## Run all tests + open coverage report for all packages
	echo 'mode: $(COVERMODE)' > .coverage
	for PKG in $(TEST_PACKAGES); do \
		go test -coverprofile=.coverage.tmp -tags "integration" $$PKG; \
		grep -v -E '^mode:' .coverage.tmp >> .coverage; \
	done
	go tool cover -html=.coverage
	$(RM) .coverage .coverage.tmp

installtools: ## Install helper tools (req for build and clean targets)
	go get -u github.com/gobuffalo/packr/v2/packr2
	go get -u github.com/swaggo/swag

build: clean build/linux build/darwin ## Build for linux and darwin (save to OUTPUT_DIR/BIN)

build/linux: clean/linux ## Build for linux (save to OUTPUT_DIR/BIN)
	GOOS=linux packr2 build -a -installsuffix cgo -ldflags $(LDFLAGS) -o $(OUTPUT_DIR)/$(BIN)-linux .

build/darwin: clean/darwin ## Build for darwin (save to OUTPUT_DIR/BIN)
	GOOS=darwin packr2 build -a -installsuffix cgo -ldflags $(LDFLAGS) -o $(OUTPUT_DIR)/$(BIN)-darwin .

clean: clean/darwin clean/linux ## Remove all build artifacts

clean/darwin: clean/packr2 ## Remove darwin build artifacts
	$(RM) $(OUTPUT_DIR)/$(BIN)-darwin

clean/linux: clean/packr2 ## Remove linux build artifacts
	$(RM) $(OUTPUT_DIR)/$(BIN)-linux

clean/packr2: ## Remove packr2 generated bits
	packr2 clean

docs: ## Generate documentation (make sure you've ran `make installtools`)
	swag init

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[34m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#'