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
# Deploying to Kubernetes
#


#  __     __         _       _     _
#  \ \   / /_ _ _ __(_) __ _| |__ | | ___ ___
#   \ \ / / _` | '__| |/ _` | '_ \| |/ _ \ __|
#    \ V / (_| | |  | | (_| | |_) | |  __\__ \
#     \_/ \__,_|_|  |_|\__,_|_.__/|_|\___|___/
#

PROJECT=$(shell gcloud config list project | awk 'FNR ==2 { print $$3 }')

#   _____                    _
#  |_   _|_ _ _ __ __ _  ___| |_ ___
#    | |/ _` | '__/ _` |/ _ \ __/ __|
#    | | (_| | | | (_| |  __/ |_\__ \
#    |_|\__,_|_|  \__, |\___|\__|___/
#                 |___/

# Deploy redis
create-redis:
	kubectl create -f redis.yaml

# create the server
create-server: server-rc create-server-service

# replica controller for the service
server-rc:
	sed 's/$${PROJECT}/$(PROJECT)/g' simon-says-rc.yaml | kubectl apply -f -

# service replica controllers
create-server-service:
	sed 's/$${PROJECT}/$(PROJECT)/g' simon-says-server.yaml | kubectl create -f -

#clear all the pods, and download them all again from the registry
restart:
	kubectl delete pods --all

#just re-deploy the server containers
update-server:
	kubectl delete pods -l role=simon-says-server
