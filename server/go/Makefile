# Copyright 2015 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Firing up the dev environment
#


#  __     __         _       _     _
#  \ \   / /_ _ _ __(_) __ _| |__ | | ___ ___
#   \ \ / / _` | '__| |/ _` | '_ \| |/ _ \ __|
#    \ V / (_| | |  | | (_| | |_) | |  __\__ \
#     \_/ \__,_|_|  |_|\__,_|_.__/|_|\___|___/
#

PROJECT?=$(shell gcloud config list project | awk 'FNR ==2 { print $$3 }')
TAG=gcr.io/$(PROJECT)/server
PACKAGE_ROOT=github.com/grpc-simonsays/simonsays-server

#Directory that this Makefile is in.
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_path := $(dir $(mkfile_path))
src_path = $(current_path)/src

export_go_path=export GOPATH=$(current_path)/vendor:$(current_path)

#   _____                    _
#  |_   _|_ _ _ __ __ _  ___| |_ ___
#    | |/ _` | '__/ _` |/ _ \ __/ __|
#    | | (_| | | | (_| |  __/ |_\__ \
#    |_|\__,_|_|  \__, |\___|\__|___/
#                 |___/

# build the docker image
build: build-static-server
	docker build --tag=$(TAG) $(current_path)

# build the entire thing, from k8s to the docker image and deploy it all.
all: build push
	cd $(current_path)/infrastructure && $(MAKE) deploy auth
	cd $(current_path)/deploy && $(MAKE) create-redis create-server

# give me a shell
shell:
	docker run -it --entrypoint=/bin/sh $(TAG)

# run it!
run:
	docker run --rm -p 50051:50051 $(TAG)

# clean all the things
clean: clean-server
	docker rmi $(TAG)

# push to gcr.io
push:
	gcloud docker push $(TAG)

# pull from gcr.io
pull:
	gcloud docker pull $(TAG)

# Build a go binary. For testing, since we need a statically compiled binary
build-server: generate-protobuf
	gb build

# Build a statically compiled binary
# https://github.com/constabulary/gb/issues/328
build-static-server: generate-protobuf
	mkdir -p $(current_path)/bin
	$(export_go_path) && \
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-w -extld ld -extldflags -static' -o $(current_path)/bin/server $(PACKAGE_ROOT)/cmd/server

# generate the go code for the protobuf (and add apache licence)
generate-protobuf: package-simonsays
	cd $(src_path) && \
	protoc --go_out=plugins=grpc:$(src_path)/$(PACKAGE_ROOT)/simonsays simonsays.proto && \
	cd ./$(PACKAGE_ROOT)/simonsays && \
	cp $(src_path)/proto_header . && \
	cat simonsays.pb.go >> proto_header && \
	mv proto_header simonsays.pb.go

# clean up the go build files, and the generated files.
clean-server:
	-rm -r $(current_path)/bin
	-rm -r $(current_path)/pkg
	-rm -r $(current_path)/vendor/pkg
	-rm $(src_path)/$(PACKAGE_ROOT)/simonsays/*.pb.go

# Make sure the code passes tests
check-code:
	-export GOPATH=$(current_path) && \
	go fmt $(PACKAGE_ROOT)/... && \
	go vet $(PACKAGE_ROOT)/... && \
	goimports -w $(src_path) && \
	golint $(PACKAGE_ROOT)/... | grep -v simonsays.pb.go

	$(export_go_path) && errcheck $(PACKAGE_ROOT)/... | grep -v simonsays.pb.go

# Fire up a godoc server
godoc:
	$(export_go_path) && \
	sleep 1 && \
	killall godoc; godoc -http=":7080" &

# make sure the simonsays package exists.
package-simonsays:
	mkdir -p $(src_path)/$(PACKAGE_ROOT)/simonsays

test:
	gb test

goconvey:
	gb build github.com/smartystreets/goconvey $(ARGS) && \
	$(export_go_path) && \
	$(current_path)/bin/goconvey -port 7081 -host 0.0.0.0
