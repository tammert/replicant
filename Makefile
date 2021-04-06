# DEV versions use Git short SHA-1, RELEASE versions use the latest Git tag.

DEV_VERSION := $(shell git rev-parse --short HEAD)
RELEASE_VERSION := $(shell git describe --tags --abbrev=0)

.PHONY: build
build:
	docker build . -t replicant:$(DEV_VERSION)

.PHONY: release
release:
	docker build . -t docker.io/tammert/replicant:$(RELEASE_VERSION) -t docker.io/tammert/replicant:latest
	docker push docker.io/tammert/replicant:$(RELEASE_VERSION)
	docker push docker.io/tammert/replicant:latest
