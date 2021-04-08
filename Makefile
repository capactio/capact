.DEFAULT_GOAL = all

# enable module support across all go commands.
export GO111MODULE = on
# enable consistent Go 1.12/1.13 GOPROXY behavior.
export GOPROXY = https://proxy.golang.org
# enable the BuildKit builder in the Docker CLI.
export DOCKER_BUILDKIT = 1

export DOCKER_REPOSITORY ?= gcr.io/projectvoltron
export DOCKER_TAG ?= latest

all: generate build-all-images test-ocf-manifests test-unit test-lint
.PHONY: all

############
# Building #
############

APPS = gateway k8s-engine och-js argo-runner helm-runner cloudsql-runner populator terraform-runner argo-actions
TESTS = e2e
INFRA = json-go-gen graphql-schema-linter jinja2

build-tool-ocftool:
	./hack/build-tool-ocftool.sh
.PHONY: build-tool-ocftool

build-tool-populator:
	./hack/build-tool-populator.sh
.PHONY: build-tool-populator

# All images
build-all-apps-images: $(addprefix build-app-image-,$(APPS))
.PHONY: build-all-apps-images

build-all-tests-images: $(addprefix build-test-image-,$(TESTS))
.PHONY: build-all-tests-images

build-all-images: build-all-apps-images build-all-tests-images $(addprefix build-infra-image-,$(INFRA))
.PHONY: build-all-images

push-all-images: $(addprefix push-app-image-,$(APPS))  $(addprefix push-test-image-,$(TESTS)) $(addprefix push-infra-image-,$(INFRA))
.PHONY: push-all-images

# App images
build-app-image-och-js:
	$(eval APP := och-js)
	cd och-js && $(MAKE) build-app-image
.PHONY: build-app-image-och-js

build-app-image-populator:
	$(eval APP := populator)
	docker build --build-arg COMPONENT=$(APP) --target generic-alpine -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-populator


build-app-image-terraform-runner:
	$(eval APP := terraform-runner)
	docker build --build-arg COMPONENT=$(APP) --target terraform-runner -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-terraform-runner

build-app-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) --target generic -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-%

push-app-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG)
.PHONY: push-apps-images-%

# Test images
build-test-image-e2e:
	docker build --build-arg COMPONENT=e2e \
		--build-arg BUILD_CMD="go test -v -c" \
		--build-arg SOURCE_PATH="./test/e2e/*_test.go" \
		--target e2e -t $(DOCKER_REPOSITORY)/e2e-test:$(DOCKER_TAG) .
.PHONY: build-test-image-e2e

build-test-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) \
		--build-arg BUILD_CMD="go test -v -c" \
		--build-arg SOURCE_PATH="./test/$(APP)/*_test.go" \
		-t $(DOCKER_REPOSITORY)/$(APP)-test:$(DOCKER_TAG) .
.PHONY: build-test-image

push-test-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/$(APP)-test:$(DOCKER_TAG)
.PHONY: push-test-image

# Infra images
INFRA_IMAGES_DIR = ./hack/images

build-infra-image-%:
	$(eval APP := $*)
	docker build -t $(DOCKER_REPOSITORY)/infra/$(APP):$(DOCKER_TAG) -f $(INFRA_IMAGES_DIR)/$(APP)/Dockerfile $(INFRA_IMAGES_DIR)/$(APP)
.PHONY: build-infra-image

push-infra-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/infra/$(APP):$(DOCKER_TAG)
.PHONY: push-infra-image

###########
# Testing #
###########

test-unit:
	./hack/test-unit.sh
.PHONY: test-unit

test-lint:
	./hack/lint.sh
.PHONY: test-lint

test-ocf-manifests:
	./hack/test-ocf-manifests.sh
.PHONY: test-ocf-manifests

test-integration:
	./hack/test-integration.sh
.PHONY: test-integration

test-k8s-controller:
	./hack/test-k8s-controller.sh
.PHONY: test-controller

test-generated:
	./hack/test-generated.sh
.PHONY: test-generated

test-cover-html: test-unit
	go tool cover -html=./coverage.txt
.PHONY: test-cover-html

##############
# Generating #
##############

generate: gen-go-api-from-ocf-spec gen-k8s-resources gen-graphql-resources gen-go-source-code gen-docs
.PHONY: generate

gen-go-api-from-ocf-spec:
	./hack/gen-go-api-from-ocf-spec.sh
.PHONY: gen-go-api

gen-k8s-resources:
	./hack/gen-k8s-resources.sh
.PHONY: gen-k8s-resources

gen-graphql-resources:
	./hack/gen-graphql-resources.sh
.PHONY: gen-graphql-resources

gen-go-source-code:
	go generate -x ./...
.PHONY: gen-go-source-code

gen-docs: gen-docs-ocftool
.PHONY: gen-docs

gen-docs-ocftool:
	go run cmd/ocftool/main.go gen-usage-docs
.PHONY: gen-docs-ocftool

###############
# Development #
###############

dev-cluster:
	./hack/dev-cluster-create.sh
.PHONY: dev-cluster

dev-cluster-update:
	./hack/dev-cluster-update.sh
.PHONY: dev-cluster-update

dev-cluster-delete:
	./hack/dev-cluster-delete.sh
.PHONY: dev-cluster-delete

fix-lint-issues:
	LINT_FORCE_FIX=true ./hack/lint.sh
.PHONY: fix-lint
