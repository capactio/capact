DOCKER_REPOSITORY ?= local
DOCKER_TAG=devel
APP=och

build-app-image:
	docker build -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-och
