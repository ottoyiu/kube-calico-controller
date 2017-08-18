PROJECT_NAME := kube-calico-controller
GOFILES:=$(shell find . -name '*.go' | grep -v -E '(./vendor)')
VERSION?=$(shell git describe --tags --dirty)
IMAGE_TAG:=ottoyiu/${PROJECT_NAME}:${VERSION}

.PHONY: all
all: clean check bin image

.PHONY: image
image:
	docker build -t ${IMAGE_TAG} .

.PHONY: push_image
push_image:
	docker push ${IMAGE_TAG}

bin: bin/linux/${PROJECT_NAME}

bin/%: LDFLAGS=-X github.com/ottoyiu/${PROJECT_NAME}.Version=${VERSION}
bin/%: $(GOFILES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=$(word 1, $(subst /, ,$*)) GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o "$@" github.com/ottoyiu/${PROJECT_NAME}/cmd/${PROJECT_NAME}/...

.PHONY: gofmt
gofmt:
	gofmt -w -s pkg/
	gofmt -w -s cmd/

.PHONY: test
test:
	go test github.com/ottoyiu/${PROJECT_NAME}/pkg/... -args -v=1 -logtostderr
	go test github.com/ottoyiu/${PROJECT_NAME}/cmd/... -args -v=1 -logtostderr

.PHONY: check
check:
	@find . -name vendor -prune -o -name '*.go' -exec gofmt -s -d {} +
	@go vet $(shell go list ./... | grep -v '/vendor/')
	@go test -v $(shell go list ./... | grep -v '/vendor/')


.PHONY: vendor
vendor:
		dep ensure

.PHONY: clean
clean:
		rm -rf bin


