DOCKER_REPOSITORY ?= ghcr.io/capactio
DOCKER_TAG ?= latest
APP ?= hub-js

build-app-image:
	docker build -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image
