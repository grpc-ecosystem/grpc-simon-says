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
RESIZE=10

#   _____                    _
#  |_   _|_ _ _ __ __ _  ___| |_ ___
#    | |/ _` | '__/ _` |/ _ \ __/ __|
#    | | (_| | | | (_| |  __/ |_\__ \
#    |_|\__,_|_|  \__, |\___|\__|___/
#                 |___/

# Deploy redis
create-redis: redis-rs redis-service

redis-rs:
	kubectl apply -f redis-rs.yaml --record

redis-service:
	kubectl create -f redis-service.yaml

# create the server
create-server: server-rs server-service

# replica set for the service
server-rs:
	@cp simon-says-rs.yaml /tmp/
	@sed -i 's/$${PROJECT}/$(PROJECT)/g' simon-says-rs.yaml
	kubectl apply -f simon-says-rs.yaml --record
	@cp /tmp/simon-says-rs.yaml .

# service replica controllers
server-service:
	kubectl create -f simon-says-service.yaml

# resize the go server cluster to 10
server-resize:
		kubectl scale deployment simon-says --replicas=$(RESIZE)

# run the visualiser
visualise:
		kubectl proxy --www=/home/$(shell whoami)/gcp-live-k8s-visualizer --www-prefix=/static/ --api-prefix=/api/ --address="0.0.0.0"

#clear all the pods, and download them all again from the registry
restart:
	kubectl delete pods --all

#just re-deploy the server containers
restart-server:
	kubectl delete pods -l role=simon-says-server

clean:
	kubectl delete deployments --all
	kubectl delete svc --all