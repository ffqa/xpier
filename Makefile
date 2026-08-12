VERSION := $(shell ./scripts/version.sh)
LDFLAGS := -X xpier/internal/xpier.Version=$(VERSION)

.PHONY: version build install test vet

version:
	@echo $(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o xpier .

install: build
	install -m 0755 xpier /usr/local/bin/xpier

test:
	go test ./...

vet:
	go vet ./...
