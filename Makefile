.DEFAULT_GOAL = all

# enable module support across all go commands.
export GO111MODULE = on
# enable consistent Go 1.12/1.13 GOPROXY behavior.
export GOPROXY = https://proxy.golang.org
# enable the BuildKit builder in the Docker CLI.
export DOCKER_BUILDKIT = 1

DOCKER_REPOSITORY ?= gcr.io/sm-cluster-dev
DOCKER_TAG ?= latest

all: generate build-all-images test-unit test-lint ## Default: generate all, build all, test all and lint
.PHONY: all

############
# Building #
############

APPS = gateway k8s-engine hub-js argo-runner helm-runner cloudsql-runner populator cli terraform-runner gcplist-runner manifestrepo-runner jira-runner argo-actions gitlab-api-runner secret-storage-backend helm-storage-backend ti-value-fetcher azure-runner autopatcher-runner capact-runner confluence-runner azviz toolbox
TESTS = e2e local-hub
INFRA = json-go-gen graphql-schema-linter jinja2 merger neo4j

# Tools
build-all-tools: ## Builds the standalone binaries for all tools
	goreleaser build --rm-dist --skip-post-hooks --snapshot --single-target
.PHONY: build-all-tools

build-tool-cli: ## Builds the standalone binaries for the capact CLI
	goreleaser build --id capact --rm-dist --skip-post-hooks --snapshot --single-target
.PHONY: build-tool-cli

build-tool-populator: ## Builds the standalone binaries for the Hub Populator
	goreleaser build --id populator --rm-dist --skip-post-hooks --snapshot --single-target
.PHONY: build-tool-populator

# All images
build-all-apps-images: $(addprefix build-app-image-,$(APPS)) ## Builds all application images
.PHONY: build-all-apps-images

build-all-tests-images: $(addprefix build-test-image-,$(TESTS)) ## Builds all test images
.PHONY: build-all-tests-images

build-all-images: build-all-apps-images build-all-tests-images $(addprefix build-infra-image-,$(INFRA)) ## Build all images
.PHONY: build-all-images

push-all-images: $(addprefix push-app-image-,$(APPS)) $(addprefix push-test-image-,$(TESTS)) $(addprefix push-infra-image-,$(INFRA)) ## Push all images to the repository
.PHONY: push-all-images

# App images
build-app-image-hub-js: ## Build application image for hub-js
	$(eval APP := hub-js)
	cd hub-js && $(MAKE) build-app-image
.PHONY: build-app-image-hub-js

build-app-image-populator: ## Build application image for Hub Populator
	$(eval APP := populator)
	docker build --build-arg COMPONENT=$(APP) --target generic-alpine -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-populator

build-app-image-terraform-runner: ## Build application image for terraform runner
	$(eval APP := terraform-runner)
	docker build --build-arg COMPONENT=$(APP) --target terraform-runner -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-terraform-runner

build-app-image-azure-runner: ## Build application image for azure runner
	$(eval APP := azure-runner)
	docker build --build-arg COMPONENT=$(APP) --target azure-runner -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-azure-runner

build-app-image-autopatcher-runner: ## Build application image for autopatcher runner
	$(eval APP := autopatcher-runner)
	docker build --build-arg COMPONENT=$(APP) --target autopatcher-runner -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-autopatcher-runner

build-app-image-cli: ## Build application image for capact cli
	$(eval APP := cli)
	docker build --build-arg COMPONENT=$(APP) --target cli -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-cli

build-app-image-capact-runner: ## Build application image for capact runner
	$(eval APP := capact-runner)
	sed -i "s/CAPACT_VERSION_REPLEACEMENT/$(DOCKER_TAG)/g" ./Dockerfile
	docker build --build-arg COMPONENT=$(APP) --target capact-runner -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-capact-runner

build-app-image-azviz:
	$(eval APP := azviz)
	docker build --build-arg COMPONENT=$(APP) --build-arg SSH_PRIVATE_KEY="$(SVG_SSH_PRIVATE_KEY)" --target azviz -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-azviz

build-app-image-toolbox:
	$(eval APP := toolbox)
	docker build --build-arg COMPONENT=$(APP) --target toolbox -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .
.PHONY: build-app-image-toolbox

build-app-image-%:
	$(eval APP := $*)
	docker build --build-arg COMPONENT=$(APP) --target generic -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) .

push-app-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG)

save-app-image-%:
	$(eval APP := $*)
	docker save $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG) > /tmp/$(APP).tar

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

push-test-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/$(APP)-test:$(DOCKER_TAG)

save-test-image-%:
	$(eval APP := $*)
	docker save $(DOCKER_REPOSITORY)/$(APP)-test:$(DOCKER_TAG) > /tmp/$(APP)-test.tar

# Infra images
INFRA_IMAGES_DIR = ./hack/images

build-infra-image-%:
	$(eval APP := $*)
	docker build -t $(DOCKER_REPOSITORY)/infra/$(APP):$(DOCKER_TAG) -f $(INFRA_IMAGES_DIR)/$(APP)/Dockerfile $(INFRA_IMAGES_DIR)/$(APP)

push-infra-image-%:
	$(eval APP := $*)
	docker push $(DOCKER_REPOSITORY)/infra/$(APP):$(DOCKER_TAG)

save-infra-image-%:
	$(eval APP := $*)
	docker save $(DOCKER_REPOSITORY)/infra/$(APP):$(DOCKER_TAG) > /tmp/infra-$(APP).tar

###########
# Testing #
###########

test-unit: ## Execute unit tests
	./hack/test-unit.sh
.PHONY: test-unit

test-lint: ## Run linters on the codebase
	./hack/lint.sh
.PHONY: test-lint

test-local-hub:
	./hack/test-local-hub.sh
.PHONY: test-local-hub

test-integration:
	./hack/test-integration.sh
.PHONY: test-integration

test-k8s-controller:
	./hack/test-k8s-controller.sh
.PHONY: test-controller

test-generated:
	./hack/test-generated.sh
.PHONY: test-generated

test-cover-html: test-unit ## Generate file with unit test coverage data
	go tool cover -html=./coverage.txt
.PHONY: test-cover-html

image-security-scan: build-all-images ## Build the docker images and check for vulnerabilities using Snyk
	./hack/scan-images.sh
.PHONY: test-image-security-scan

##############
# Generating #
##############

generate: gen-go-api-from-ocf-spec gen-k8s-resources gen-graphql-resources gen-go-source-code gen-docs gen-grpc-resources ## Run all generators
.PHONY: generate

gen-go-api-from-ocf-spec: ## Generate Go code from OCF JSON Schemas
	./hack/gen-go-api-from-ocf-spec.sh
.PHONY: gen-go-api

gen-k8s-resources: ## Generate K8s resources
	./hack/gen-k8s-resources.sh
.PHONY: gen-k8s-resources

gen-graphql-resources: ## Generate code from GraphQL schema
	./hack/gen-graphql-resources.sh
.PHONY: gen-graphql-resources

gen-grpc-resources: ## Generate gRPC + ProtoBuf Go code for client and server
	./hack/gen-grpc-resources.sh
.PHONY: gen-proto-source-code

gen-go-source-code:
	go generate -x ./...
.PHONY: gen-go-source-code

gen-docs: gen-docs-cli gen-docs-populator ## Generate all documentation
.PHONY: gen-docs

gen-docs-cli:
	rm -f ./cmd/cli/docs/*
	go run cmd/cli/main.go gen-usage-docs
.PHONY: gen-docs-cli

gen-docs-populator:
	rm -f ./cmd/populator/docs/*
	go run cmd/populator/main.go gen-usage-docs
.PHONY: gen-docs-populator

###############
# Development #
###############

dev-cluster: ## Create the dev cluster
	./hack/dev-cluster-create.sh
.PHONY: dev-cluster

dev-cluster-update: ## Updadte the dev cluster
	./hack/dev-cluster-update.sh
.PHONY: dev-cluster-update

dev-cluster-delete: ## Delete the dev cluster
	./hack/dev-cluster-delete.sh
.PHONY: dev-cluster-delete

fix-lint-issues: ## Automatically fix lint issues
	LINT_FORCE_FIX=true ./hack/lint.sh
.PHONY: fix-lint

#############
# Releasing #
#############

release-charts: ## Release Capact Helm Charts
	./hack/release-charts.sh
.PHONY: release-charts

release-binaries: ## Release stable Capact binaries, such as CLI, populator etc.
	# --skip-validate is needed as we need to locally ignore GCP credentials
	# by creating a commit, which is not pushed to the origin.
	# That makes the goreleaser's validation fail with error "it tag {tag} was not made against commit {commit hash}"
	# Even if we not create such commit, goreleaser would fail because of git dirty state.
	goreleaser release --skip-validate --rm-dist --timeout 60m

release-latest-binaries: ## Release latest Capact binaries
	goreleaser release --snapshot --rm-dist --config .goreleaser.latest.yml
	# Needs to be published manually, as gorelaser doesn't support latest tag
	# https://goreleaser.com/limitations/semver/
	#
	# Update binaries
	gsutil -m rsync -x "goreleaserdocker.*" -r ./bin/ gs://supermaestro-binaries/latest/
	# By default Google sets `cache-control: public, max-age=3600`.
	# We need to change to ensure the file is not cached by http clients, so latest version is always downloaded
	# source: https://cloud.google.com/storage/docs/caching#performance_considerations
	gsutil setmeta -h "Cache-Control: no-cache, no-store" gs://supermaestro-binaries/latest/*
	# Update Docker images
	docker push gcr.io/sm-cluster-dev/tools/capact-cli:latest

#############
# Other     #
#############

clean: ## Cleans all files/directories defined in .gitignore
	git clean
.PHONY: clean

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
.PHONY: help
