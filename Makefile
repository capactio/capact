.DEFAULT_GOAL = all

# enable module support across all go commands.
export GO111MODULE = on
# enable consistent Go 1.12/1.13 GOPROXY behavior.
export GOPROXY = https://proxy.golang.org

DOCKER_PUSH_REPOSITORY ?= gcr.io/projectvoltron/
DOCKER_TAG ?= latest

all: build-all-images test-unit test-lint
.PHONY: all

#
# Build components
#
APPS = gateway k8s-engine och
TESTS = e2e

# All images
build-all-images: $(addprefix build-app-image-,$(APPS)) $(addprefix build-test-image-,$(TESTS))
.PHONY: build-all-images

push-all-images: $(addprefix push-app-image-,$(APPS)) $(addprefix push-test-image-,$(TESTS))
.PHONY: push-all-images

# App images
build-app-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) -t $(DOCKER_PUSH_REPOSITORY)$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-%

push-app-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_PUSH_REPOSITORY)$(APP):$(DOCKER_TAG)
.PHONY: push-apps-images-%

# Test images
build-test-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) \
		--build-arg BUILD_CMD="go test -v -c" \
		--build-arg SOURCE_PATH="./test/$(APP)/$(APP)_test.go" \
		-t $(DOCKER_PUSH_REPOSITORY)$(APP)_test:$(DOCKER_TAG) .
.PHONY: build-test-image

push-test-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_PUSH_REPOSITORY)$(APP)_test:$(DOCKER_TAG)
.PHONY: push-test-image

#
# Test
#
test-unit:
	./hack/run-test-unit.sh
.PHONY: test-unit

test-lint:
	./hack/run-lint.sh
.PHONY: test-lint

cover-html: test-unit
	go tool cover -html=./coverage.txt
.PHONY: cover-html
