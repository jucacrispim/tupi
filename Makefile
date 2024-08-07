GOCMD=go
GOBUILD=$(GOCMD) build -trimpath
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v ./... -trimpath
BIN_NAME=tupi
BUILD_DIR=build
BIN_PATH=./$(BUILD_DIR)/$(BIN_NAME)
OUTFLAG=-o $(BIN_PATH)
CMDFILE=cmd/tupi/main.go
TESTDATA_DIR=testdata
PLUGIN_MODE_FLAG=-buildmode=plugin

AUTH_PLUGIN_BIN_PATH=./$(BUILD_DIR)/auth_plugin.so
BAD_AUTH_PLUGIN_BIN_PATH=./$(BUILD_DIR)/auth_plugin_bad.so
PANIC_AUTH_PLUGIN_BIN_PATH=./$(BUILD_DIR)/auth_plugin_panic.so

SERVE_PLUGIN_BIN_PATH=./$(BUILD_DIR)/serve_plugin.so
BAD_SERVE_PLUGIN_BIN_PATH=./$(BUILD_DIR)/serve_plugin_bad.so

INIT_PLUGIN_BIN_PATH=./$(BUILD_DIR)/init_plugin.so
BAD_INIT_PLUGIN_BIN_PATH=./$(BUILD_DIR)/init_plugin_bad.so
PANIC_INIT_PLUGIN_BIN_PATH=./$(BUILD_DIR)/init_plugin_panic.so

AUTH_PLUGIN_FILE=$(TESTDATA_DIR)/auth_plugin.go
BAD_AUTH_PLUGIN_FILE=$(TESTDATA_DIR)/auth_plugin_bad.go
PANIC_AUTH_PLUGIN_FILE=$(TESTDATA_DIR)/auth_plugin_panic.go
SERVE_PLUGIN_FILE=$(TESTDATA_DIR)/serve_plugin.go
BAD_SERVE_PLUGIN_FILE=$(TESTDATA_DIR)/serve_plugin_bad.go

INIT_PLUGIN_FILE=$(TESTDATA_DIR)/init_plugin.go
BAD_INIT_PLUGIN_FILE=$(TESTDATA_DIR)/init_plugin_bad.go
PANIC_INIT_PLUGIN_FILE=$(TESTDATA_DIR)/init_plugin_panic.go



.PHONY: build # - Creates the binary under the build/ directory
build:
	$(GOBUILD) $(OUTFLAG) $(CMDFILE)

.PHONY: buildtest # - Creates the binary for the test plugins under the build/ directory
buildtest:
	$(GOBUILD) -o $(AUTH_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(AUTH_PLUGIN_FILE)
	$(GOBUILD) -o $(BAD_AUTH_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(BAD_AUTH_PLUGIN_FILE)
	$(GOBUILD) -o $(PANIC_AUTH_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(PANIC_AUTH_PLUGIN_FILE)

	$(GOBUILD) -o $(INIT_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(INIT_PLUGIN_FILE)
	$(GOBUILD) -o $(BAD_INIT_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(BAD_INIT_PLUGIN_FILE)
	$(GOBUILD) -o $(PANIC_INIT_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(PANIC_INIT_PLUGIN_FILE)

	$(GOBUILD) -o $(SERVE_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(SERVE_PLUGIN_FILE)
	$(GOBUILD) -o $(BAD_SERVE_PLUGIN_BIN_PATH) $(PLUGIN_MODE_FLAG) $(BAD_SERVE_PLUGIN_FILE)


.PHONY: test # - Run all tests
test:
	$(GOBUILD)
	$(GOTEST)

.PHONY: setupenv # - Install needed tools for tests
setupenv:
	./build-scripts/env.sh setup-env

.PHONY: setupenv # - Install needed tools for docs
setupdocsenv:
	./build-scripts/env.sh setup-docs-env


.PHONY: docs # - Build documentation
docs:
	./build-scripts/env.sh build-docs

.PHONY: cov # - Run all tests and check coverage
cov: buildtest coverage

coverage:
	./build-scripts/check_coverage.sh

.PHONY: run # - Run the program. You can use `make run ARGS="-host :9090 -root=/"`
run:
	$(GOBUILD) $(OUTFLAG)
	$(BIN_PATH) $(ARGS)

.PHONY: clean # - Remove the files created during build
clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache

.PHONY: install # - Copy the binary to the path
install: build
	go install -trimpath github.com/jucacrispim/tupi/cmd/tupi

.PHONY: uninstall # - Remove the binary from path
uninstall:
	go clean -i github.com/jucacrispim/tupi


.PHONY: help  # - Show this help text
help:
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1 \2/' | expand -t20

all: build test install
