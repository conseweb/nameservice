# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#
# -------------------------------------------------------------
# This makefile defines the following targets
#
#   - all (default) - builds all targets and runs all tests/checks
#   - checks - runs all tests/checks
#   - peer - builds the fabric peer binary
#   - membersrvc - builds the membersrvc binary
#   - unit-test - runs the go-test based unit tests
#   - behave - runs the behave test
#   - behave-deps - ensures pre-requisites are availble for running behave manually
#   - gotools - installs go tools like golint
#   - linter - runs all code checks
#   - images[-clean] - ensures all docker images are available[/cleaned]
#   - peer-image[-clean] - ensures the peer-image is available[/cleaned] (for behave, etc)
#   - membersrvc-image[-clean] - ensures the membersrvc-image is available[/cleaned] (for behave, etc)
#   - protos - generate all protobuf artifacts based on .proto files
#   - node-sdk - builds the node.js client sdk
#   - node-sdk-unit-tests - runs the node.js client sdk unit tests
#   - clean - cleans the build area
#   - dist-clean - superset of 'clean' that also removes persistent state

PROJECT_NAME   = hyperledger/fabric
BASE_VERSION   = 0.6.2-preview
IS_RELEASE     = false

ifneq ($(IS_RELEASE),true)
EXTRA_VERSION ?= snapshot-$(shell git rev-parse --short HEAD)
PROJECT_VERSION=$(BASE_VERSION)-$(EXTRA_VERSION)
else
PROJECT_VERSION=$(BASE_VERSION)
endif

DOCKER_TAG=$(shell uname -m)-$(PROJECT_VERSION)

PKGNAME = github.com/$(PROJECT_NAME)
GO_LDFLAGS = -X github.com/hyperledger/fabric/metadata.Version=$(PROJECT_VERSION)
CGO_FLAGS = CGO_CFLAGS=" " CGO_LDFLAGS="-lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy"
UID = $(shell id -u)
CHAINTOOL_RELEASE=v0.9.1

# EXECUTABLES = go docker git curl
# K := $(foreach exec,$(EXECUTABLES),\
# 	$(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH: Check dependencies")))

# SUBDIRS are components that have their own Makefiles that we can invoke
SUBDIRS = gotools sdk/node
SUBDIRS:=$(strip $(SUBDIRS))

# Make our baseimage depend on any changes to images/base or scripts/provision
BASEIMAGE_RELEASE = $(shell cat ./images/base/release)
BASEIMAGE_DEPS    = $(shell git ls-files images/base scripts/provision)

JAVASHIM_DEPS =  $(shell git ls-files core/chaincode/shim/java)
PROJECT_FILES = $(shell git ls-files)
IMAGES = base src ccenv peer membersrvc javaenv

PWD := $(shell pwd)
IMAGE := ckeyer/obc:dev
INNER_GOPATH := /opt/gopath
NET := $(shell docker network inspect cknet > /dev/zero && echo "--net cknet --ip 172.16.1.5" || echo "")

ASSET := vue
DIST_URL := "http://ckeyer:Nzf6tDiWLwEuqW2krQDd@d.cj0.pw/farmer-ui-$(ASSET).tgz"

GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
IMAGE_NAME := hub.conseweb.com:5000/farmer:$(GIT_BRANCH)

assets:
	curl -sS $(DIST_URL)|gzip -dc|tar x
	cd dist && go-bindata -o ../farmer/api/views/assets.go -pkg=views ./...
	rm -rf dist


all: peer membersrvc checks

checks: linter unit-test behave

.PHONY: $(SUBDIRS)
$(SUBDIRS):
	cd $@ && $(MAKE)

.PHONY: peer
peer: build/bin/peer
peer-image: build/image/peer/.dummy

.PHONY: membersrvc
membersrvc: build/bin/membersrvc
membersrvc-image: build/image/membersrvc/.dummy

unit-test: peer-image gotools
	@./scripts/goUnitTests.sh $(DOCKER_TAG) "$(GO_LDFLAGS)"

node-sdk: sdk/node

node-sdk-unit-tests: peer membersrvc
	cd sdk/node && $(MAKE) unit-tests

unit-tests: unit-test node-sdk-unit-tests

.PHONY: images
images: $(patsubst %,build/image/%/.dummy, $(IMAGES))

behave-deps: images peer build/bin/block-listener
behave: behave-deps
	@echo "Running behave tests"
	@cd bddtests; behave $(BEHAVE_OPTS)

linter: gotools
	@echo "LINT: Running code checks.."
	@echo "Running go vet"
	go vet ./consensus/...
	go vet ./core/...
	go vet ./events/...
	go vet ./examples/...
	go vet ./membersrvc/...
	go vet ./peer/...
	go vet ./protos/...
	@echo "Running goimports"
	@./scripts/goimports.sh

# We (re)build protoc-gen-go from within docker context so that
# we may later inject the binary into a different docker environment
# This is necessary since we cannot guarantee that binaries built
# on the host natively will be compatible with the docker env.
%/bin/protoc-gen-go: build/image/base/.dummy Makefile
	@echo "Building $@"
	@mkdir -p $(@D)
	@docker run -i \
		--user=$(UID) \
		-v $(abspath vendor/github.com/golang/protobuf):/opt/gopath/src/github.com/golang/protobuf \
		-v $(abspath $(@D)):/opt/gopath/bin \
		hyperledger/fabric-baseimage go install github.com/golang/protobuf/protoc-gen-go

build/bin/chaintool: Makefile
	@echo "Installing chaintool"
	@mkdir -p $(@D)
	curl -L https://github.com/hyperledger/fabric-chaintool/releases/download/$(CHAINTOOL_RELEASE)/chaintool > $@
	chmod +x $@

%/bin/chaintool: build/bin/chaintool
	@mkdir -p $(@D)
	@cp $^ $@

# JIRA FAB-243 - Mark build/docker/bin artifacts explicitly as secondary
#                since they are never referred to directly. This prevents
#                the makefile from deleting them inadvertently.
.SECONDARY: build/docker/bin/peer build/docker/bin/membersrvc

# We (re)build a package within a docker context but persist the $GOPATH/pkg
# directory so that subsequent builds are faster
build/docker/bin/%: build/image/src/.dummy $(PROJECT_FILES)
	$(eval TARGET = ${patsubst build/docker/bin/%,%,${@}})
	@echo "Building $@"
	@mkdir -p build/docker/bin build/docker/pkg
	@docker run -i \
		--user=$(UID) \
		-v $(abspath build/docker/bin):/opt/gopath/bin \
		-v $(abspath build/docker/pkg):/opt/gopath/pkg \
		hyperledger/fabric-src:$(DOCKER_TAG) go install -ldflags "$(GO_LDFLAGS)" github.com/hyperledger/fabric/$(TARGET)
	@touch $@

build/bin:
	mkdir -p $@

# Both peer and peer-image depend on ccenv-image and javaenv-image (all docker env images it supports)
build/bin/peer: build/image/ccenv/.dummy build/image/javaenv/.dummy
build/image/peer/.dummy: build/image/ccenv/.dummy build/image/javaenv/.dummy

build/bin/block-listener:
	@mkdir -p $(@D)
	$(CGO_FLAGS) GOBIN=$(abspath $(@D)) go install $(PKGNAME)/examples/events/block-listener
	@echo "Binary available as $@"
	@touch $@

build/bin/%: build/image/base/.dummy $(PROJECT_FILES)
	@mkdir -p $(@D)
	@echo "$@"
	$(CGO_FLAGS) GOBIN=$(abspath $(@D)) go install -ldflags "$(GO_LDFLAGS)" $(PKGNAME)/$(@F)
	@echo "Binary available as $@"
	@touch $@

# Special override for base-image.
build/image/base/.dummy: $(BASEIMAGE_DEPS)
	@echo "Building docker base-image"
	@mkdir -p $(@D)
	@./scripts/provision/docker.sh $(BASEIMAGE_RELEASE)
	@touch $@

# Special override for src-image
build/image/src/.dummy: build/image/base/.dummy $(PROJECT_FILES)
	@echo "Building docker src-image"
	@mkdir -p $(@D)
	@cat images/src/Dockerfile.in \
		| sed -e 's/_TAG_/$(DOCKER_TAG)/g' \
		> $(@D)/Dockerfile
	@git ls-files | tar -jcT - > $(@D)/gopath.tar.bz2
	docker build -t $(PROJECT_NAME)-src $(@D)
	docker tag $(PROJECT_NAME)-src $(PROJECT_NAME)-src:$(DOCKER_TAG)
	@touch $@

# Special override for ccenv-image (chaincode-environment)
build/image/ccenv/.dummy: build/image/src/.dummy build/image/ccenv/bin/protoc-gen-go build/image/ccenv/bin/chaintool Makefile
	@echo "Building docker ccenv-image"
	@cat images/ccenv/Dockerfile.in \
		| sed -e 's/_TAG_/$(DOCKER_TAG)/g' \
		> $(@D)/Dockerfile
	docker build -t $(PROJECT_NAME)-ccenv $(@D)
	docker tag $(PROJECT_NAME)-ccenv $(PROJECT_NAME)-ccenv:$(DOCKER_TAG)
	@touch $@

# Special override for java-image
# Following items are packed and sent to docker context while building image
# 1. Java shim layer source code
# 2. Proto files used to generate java classes
# 3. Gradle settings file
build/image/javaenv/.dummy: Makefile $(JAVASHIM_DEPS)
	@echo "Building docker javaenv-image"
	@mkdir -p $(@D)
	@cat images/javaenv/Dockerfile.in > $(@D)/Dockerfile
	@git ls-files core/chaincode/shim/java | tar -jcT - > $(@D)/javashimsrc.tar.bz2
	@git ls-files protos core/chaincode/shim/table.proto settings.gradle  | tar -jcT - > $(@D)/protos.tar.bz2
	docker build -t $(PROJECT_NAME)-javaenv $(@D)
	docker tag $(PROJECT_NAME)-javaenv $(PROJECT_NAME)-javaenv:$(DOCKER_TAG)
	@touch $@

# Default rule for image creation
build/image/%/.dummy: build/image/src/.dummy build/docker/bin/%
	$(eval TARGET = ${patsubst build/image/%/.dummy,%,${@}})
	@echo "Building docker $(TARGET)-image"
	@mkdir -p $(@D)/bin
	@cat images/app/Dockerfile.in \
		| sed -e 's/_TAG_/$(DOCKER_TAG)/g' \
		> $(@D)/Dockerfile
	cp build/docker/bin/$(TARGET) $(@D)/bin
	docker build -t $(PROJECT_NAME)-$(TARGET) $(@D)
	docker tag $(PROJECT_NAME)-$(TARGET) $(PROJECT_NAME)-$(TARGET):$(DOCKER_TAG)
	@touch $@

.PHONY: protos
protos: gotools
	./devenv/compile_protos.sh

base-image-clean:
	-docker rmi -f $(PROJECT_NAME)-baseimage
	-@rm -rf build/image/base ||:

src-image-clean: ccenv-image-clean peer-image-clean membersrvc-image-clean

%-image-clean:
	$(eval TARGET = ${patsubst %-image-clean,%,${@}})
	-docker images -q $(PROJECT_NAME)-$(TARGET) | xargs -r docker rmi -f
	-@rm -rf build/image/$(TARGET) ||:

images-clean: $(patsubst %,%-image-clean, $(IMAGES))

.PHONY: $(SUBDIRS:=-clean)
$(SUBDIRS:=-clean):
	cd $(patsubst %-clean,%,$@) && $(MAKE) clean

.PHONY: clean
clean: images-clean $(filter-out gotools-clean, $(SUBDIRS:=-clean))
	-@rm -rf build ||:

.PHONY: dist-clean
dist-clean: clean gotools-clean
	-@rm -rf /var/hyperledger/* ||:

test: test-indocker

test-indocker:
	docker run --rm \
	 --name farmer-testing \
	 -v $(PWD):$(INNER_GOPATH)/src/$(PKGNAME) \
	 -w $(INNER_GOPATH)/src/$(PKGNAME) \
	 -v /var/run/docker.sock:/var/run/docker.sock \
	 -i $(IMAGE) make test-local

test-local: assets
	go test $$(go list ./...|grep -v "vendor"|grep -v "examples"|grep -v "java"|grep -v "platforms"|grep -v "chaincode"|grep -v "core/comm")

dev:
	docker run --rm \
	 $(NET) \
	 -p 9375:9375 \
	 -p 7050:7050 \
	 -p 7051:7051 \
	 -p 7052:7052 \
	 -p 7053:7053 \
	 -p 7054:7054 \
	 --name dev \
	 -v $(PWD):$(INNER_GOPATH)/src/$(PKGNAME) \
	 -v $(GOPATH)/src/github.com/conseweb/common:$(INNER_GOPATH)/src/github.com/conseweb/common \
	 -w $(INNER_GOPATH)/src/$(PKGNAME) \
	 -v /var/run/docker.sock:/var/run/docker.sock \
	 -v /data:/data \
	 -it $(IMAGE) bash

local: local-build daemon

local-build:
	go build -o bin/farmer ./peer/main.go

build-inDocker:
	docker run --rm \
	 --name farmer-building \
	 -v $(PWD):$(INNER_GOPATH)/src/$(PKGNAME) \
	 -w $(INNER_GOPATH)/src/$(PKGNAME) \
	 $(IMAGE) make local-build

build-image: #build-inDocker
	-rm -rf ./build
	mkdir build
	cp bin/farmer build/
	cp -a $(GOPATH)/src/github.com/conseweb/common build/
	docker build -t $(IMAGE_NAME) -f Dockerfile.run .

daemon: 
	-rm /data/hyperledger/farmer.pid
	./bin/farmer farmer start

clean-runing-file:
	rm -rf /data/hyperledger/

BIN := ./bin/farmer
CHAINCODE := a4700768debee50f069102d5be41310d3852e29a544ec3cc010d6b0bf577dfa16ffb627677793dce976ec9e9e76d73467672aba8c5c31517a285e2e66b60a547
deploy:
	$(BIN) chaincode deploy -p github.com/conseweb/common/assets/lepuscoin -c '{"Function":"deployfunc", "Args": ["aaa"]}'

invoke:
	$(BIN) chaincode invoke -n $(CHAINCODE) -c '{"Function":"invoke_coinbase", "Args": [ ]}'

invoke2:
	$(BIN) chaincode invoke -n $(CHAINCODE) -c '{"Function":"invoke_transfer", "Args": [ ]}'
 