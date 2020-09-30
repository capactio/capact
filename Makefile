.DEFAULT_GOAL=all

# enable module support across all go commands.
export GO111MODULE=on
# enable consistent Go 1.12/1.13 GOPROXY behavior.
export GOPROXY=https://proxy.golang.org

all: build-all test-unit test-lint
.PHONY: all

# Build components
ALL_ARCH=linux darwin
CMD=gateway k8s-engine och

build-all: $(addprefix build-for-,$(ALL_ARCH))
.PHONY: build-all

build-for-%:
	$(MAKE) ARCH=$* build

build:
	for command in ${CMD}; do \
		CGO_ENABLED=0 GOOS=$(ARCH) GOARCH=amd64 go build -ldflags "-s -w" -o "bin/$(ARCH)/$${command}" "cmd/$${command}/main.go"; \
	done; \
	CGO_ENABLED=0 GOOS=$(ARCH) GOARCH=amd64 go test -v -c -o "bin/$(ARCH)/e2e_test" ./test/e2e/e2e_test.go;
.PHONY: build


# Test
test-unit:
	./hack/run-test-unit.sh
.PHONY: test-unit

test-lint:
	./hack/run-lint.sh
.PHONY: test-lint

cover-html: test-unit
	go tool cover -html=./coverage.txt
.PHONY: cover-html
