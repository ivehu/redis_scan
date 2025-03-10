GO=go
BINARY=redis_scan
VERSION=1.0.0

.PHONY: all build install clean

all: build

build:
	@echo "Building $(BINARY)..."
	${GO} build -o ${BINARY} -ldflags "-X main.Version=${VERSION}" .

install:
	@echo "Installing to /usr/local/bin..."
	cp ${BINARY} /usr/local/bin/

clean:
	@echo "Cleaning..."
	rm -f ${BINARY}

run:
	@${GO} run .

test:
	@${GO} test -v ./...

deps:
	${GO} mod download