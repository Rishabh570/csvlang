BIN_NAME = csvlang

.PHONY: build
build:
	goreleaser build --single-target --clean

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean

.PHONY: release
release:
	goreleaser release --clean
