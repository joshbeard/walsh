EXEC_NAME=walsh
PACKAGE_NAME=walsh

MODULE_NAME=$(shell go list -m)

DIST_DIR=dist

# Used for packaging
VERSION?=$(shell git describe --tags --always)

LD_FLAGS="-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(shell git rev-parse --short HEAD)' \
	-X 'main.date=$(shell date -u '+%Y-%m-%d')'"

# Golang-CI Lint image used for linting.
GOLANGCI_LINT_DOCKER_IMAGE := golangci/golangci-lint:latest-alpine

ifeq ($(shell uname),Linux)
        CHECKSUM_CMD = sha256sum
endif
ifeq ($(shell uname),Darwin)
        CHECKSUM_CMD = shasum -a 256 -b
endif

# Run 'make help' for a list of targets.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Shows this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Linting ##
.PHONY: vet
vet: ## Runs 'go vet'
	go vet ./...

.PHONY: gofumpt
gofumpt: vet ## Check linting with 'gofumpt'
	go run mvdan.cc/gofumpt -l -d .

.PHONY: lines
lines: ## Check long lines.
	golines -m 100 --dry-run ./

.PHONY: lines-fix
lines-fix: lines ## Fix long lines.
	golines -m 100 -w ./

.PHONY: lint
lint: vet gofumpt lines ## Lint using 'golangci-lint'
	go run github.com/golangci/golangci-lint/cmd/golangci-lint \
		run --timeout=300s ./...

.PHONY: lint-ci
lint-ci: vet gofumpt lines ## Lint using 'golangci-lint'
	go run github.com/golangci/golangci-lint/cmd/golangci-lint \
		run --timeout=300s --out-format checkstyle ./... 2>&1 | tee checkstyle-report.xml

## Testing ##
.PHONY: test
test: ## Run unit and race tests with 'go test'
	go test -count=1 -parallel=4 -coverprofile=coverage.txt -covermode count -coverpkg=./...
	go test -race -short ./...

## Coverage ##
.PHONY: coverage
coverage: test ## Generate a code test coverage report using 'gocover-cobertura'
	go run github.com/boumenot/gocover-cobertura < coverage.txt > coverage.xml
	rm -f coverage.txt

.PHONY: covopen
covopen: test ## Open the coverage report in a browser
	go tool cover -html=coverage.txt

.PHONY: testall
testall: lint test coverage ## Run linting, tests, and generates a coverage report

## Vulnerability checks ##
.PHONY: check-vuln
check-vuln: ## Check for vulnerabilities using 'govulncheck'
	@echo "Checking for vulnerabilities..."
	go run golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: build
build: ## Build the binary for the host platform
	mkdir -p dist
	go build -ldflags $(LD_FLAGS) -o $(DIST_DIR)/$(EXEC_NAME)

.PHONY: install
install: ## Install the binary
	go install -ldflags $(LD_FLAGS)

.PHONY: clean
clean: ## Clean test files
	rm -f coverage.txt coverage.xml coverage.html checkstyle-report.xml
