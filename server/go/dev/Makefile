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
# Makefile for the server development environment
#


#  __     __         _       _     _
#  \ \   / /_ _ _ __(_) __ _| |__ | | ___ ___
#   \ \ / / _` | '__| |/ _` | '_ \| |/ _ \ __|
#    \ V / (_| | |  | | (_| | |_) | |  __\__ \
#     \_/ \__,_|_|  |_|\__,_|_.__/|_|\___|___/
#

PROJECT?=$(shell gcloud config list project | awk 'FNR ==2 { print $$3 }')
TAG=gcr.io/$(PROJECT)/server-dev
NAME=simon-says-server-shell
DOCKER_GID=$(word 3,$(subst :, ,$(shell getent group docker)))
PORT=7080

#Directory that this Makefile is in.
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_path := $(dir $(mkfile_path))
project_path := $(realpath $(current_path)/..)
root_path := $(realpath $(project_path)/..)

#   _____                    _
#  |_   _|_ _ _ __ __ _  ___| |_ ___
#    | |/ _` | '__/ _` |/ _ \ __/ __|
#    | | (_| | | | (_| |  __/ |_\__ \
#    |_|\__,_|_|  \__, |\___|\__|___/
#                 |___/

# build the docker dev image
build:
	docker build --tag=$(TAG) $(current_path)

push:
	gcloud docker push $(TAG)

pull:
	gcloud docker pull $(TAG)

#clean up the shell
clean:
	docker rmi $(TAG)

# Start a dev shell
shell:
	mkdir -p ~/.config/gcloud
	mkdir -p $(project_path)/.kube
	docker run --rm \
		--name=$(NAME) \
		-p=8001:8001 \
		-P=true \
		-e TERM \
		-e HOST_GID=`id -g` \
		-e HOST_UID=`id -u` \
		-e HOST_USER=$(USER) \
		-e DOCKER_GID=$(DOCKER_GID) \
		-v ~/.config/gcloud:/home/$(USER)/.config/gcloud \
		-v ~/.appcfg_oauth2_tokens:/home/$(USER)/.appcfg_oauth2_tokens \
		-v $(project_path)/.kube:/home/$(USER)/.kube \
		-v $(project_path)/dev/zshrc:/home/$(USER)/.zshrc \
		-v /usr/bin/docker:/usr/bin/docker \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(project_path):/home/$(USER)/server \
		-w /home/$(USER)/server \
		-it $(TAG) /root/startup.sh

# attach to a running dev shell
shell-attach:
	docker exec -it --user=$(USER) $(NAME) zsh

# mount the go sdk
mount-go:
	mkdir -p /tmp/$(NAME)/go
	sshfs $(USER)@0.0.0.0:/usr/local/go /tmp/$(NAME)/go -p $(call getPort,22) -o follow_symlinks

# open chrome to the forwarded port
chrome:
	google-chrome http://localhost:$(call getPort,$(PORT))

#   _____                 _   _
#  |  ___|   _ _ __   ___| |_(_) ___  _ __  ___
#  | |_ | | | | '_ \ / __| __| |/ _ \| '_ \/ __|
#  |  _|| |_| | | | | (__| |_| | (_) | | | \__ \
#  |_|   \__,_|_| |_|\___|\__|_|\___/|_| |_|___/
#

# get the mapped docker host port
getPort = $(word 2,$(subst :, ,$(shell docker port $(NAME) $(1))))
