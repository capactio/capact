.DEFAULT_GOAL = all

# enable module support across all go commands.
export GO111MODULE = on
# enable consistent Go 1.12/1.13 GOPROXY behavior.
export GOPROXY = https://proxy.golang.org

DOCKER_PUSH_REPOSITORY ?= gcr.io/projectvoltron
DOCKER_TAG ?= latest

all: generate build-all-images test-spec test-unit test-lint
.PHONY: all

############
# Building #
############

APPS = gateway k8s-engine och
TESTS = e2e
INFRA = json-go-gen

# All images
build-all-images: $(addprefix build-app-image-,$(APPS)) $(addprefix build-test-image-,$(TESTS)) $(addprefix build-infra-image-,$(INFRA))
.PHONY: build-all-images

push-all-images: $(addprefix push-app-image-,$(APPS))  $(addprefix push-test-image-,$(TESTS)) $(addprefix push-infra-image-,$(INFRA))
.PHONY: push-all-images

# App images
build-app-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) -t $(DOCKER_PUSH_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-%

push-app-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_PUSH_REPOSITORY)/$(APP):$(DOCKER_TAG)
.PHONY: push-apps-images-%

# Test images
build-test-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) \
		--build-arg BUILD_CMD="go test -v -c" \
		--build-arg SOURCE_PATH="./test/$(APP)/$(APP)_test.go" \
		-t $(DOCKER_PUSH_REPOSITORY)/$(APP)-test:$(DOCKER_TAG) .
.PHONY: build-test-image

push-test-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_PUSH_REPOSITORY)/$(APP)-test:$(DOCKER_TAG)
.PHONY: push-test-image

# Infra images
build-infra-image-%:
	$(eval APP := $*)
	docker build -t $(DOCKER_PUSH_REPOSITORY)/infra/$(APP):$(DOCKER_TAG) -f ./hack/images/$(APP)/Dockerfile .
.PHONY: build-infra-image

push-infra-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_PUSH_REPOSITORY)/infra/$(APP):$(DOCKER_TAG)
.PHONY: push-infra-image

###########
# Testing #
###########

test-unit:
	./hack/run-test-unit.sh
.PHONY: test-unit

test-lint:
	./hack/run-lint.sh
.PHONY: test-lint

test-spec:
	go test -v --tags=ocfexamples ocf-spec/0.0.1/examples/examples_test.go
.PHONY: test-spec

test-integration:
	./hack/run-test-integration.sh
.PHONY: test-integration

test-k8s-controller:
	./hack/run-test-k8s-controller.sh
.PHONY: test-controller

test-cover-html: test-unit
	go tool cover -html=./coverage.txt
.PHONY: test-cover-html

##############
# Generating #
##############

generate: gen-go-api-from-ocf-spec gen-k8s-resources
.PHONY: generate

gen-go-api-from-ocf-spec:
	./hack/gen-go-api-from-ocf-spec.sh
.PHONY: gen-go-api

gen-k8s-resources:
	./hack/gen-k8s-resources.sh
.PHONY: gen-k8s-resources

###############
# Development #
###############
dev-cluster:
	./hack/run-dev-cluster.sh
.PHONY: dev-cluster

dev-cluster-update:
	./hack/run-dev-cluster-update.sh
.PHONY: dev-cluster-update

fix-lint-issues:
	LINT_FORCE_FIX=true ./hack/run-lint.sh
.PHONY: fix-lint
