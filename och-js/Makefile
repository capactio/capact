DOCKER_REPOSITORY ?= gcr.io/projectvoltron
DOCKER_TAG ?= latest
APP ?= och-js

build-app-image:
	docker build -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image
