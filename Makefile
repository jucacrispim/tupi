GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BIN_NAME=tupi
BUILD_DIR=build
BIN_PATH=./$(BUILD_DIR)/$(BIN_NAME)
OUTFLAG=-o $(BIN_PATH)


.PHONY: build # - Create the binary under the build/ directory
build:
	$(GOBUILD) $(OUTFLAG)

.PHONY: test # - Run all tests
test:
	go test -v ./...


.PHONY: run # - Run the program. You can use `make run ARGS="-host :9090 -root=/"`
run:
	$(GOBUILD) $(OUTFLAG)
	$(BIN_PATH) $(ARGS)

.PHONY: clean # - Remove the files created during build
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install # - Copy the binary to the path
install:
	$(GOBUILD) $(OUTFLAG)
	cp $(BIN_PATH) .
	go install

.PHONY: uninstall # - Remove the binary from path
uninstall:
	go clean -i github.com/jucacrispim/tupi


.PHONY: help  # - Show help text
help:
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1 \2/' | expand -t20

all: build test install
