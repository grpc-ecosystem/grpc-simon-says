# Deploying to the Kubernetes Cluster

Deploying containersa to the Kubernetes cluster

## Deploying Redis

To deploy the Redis container, and service, run `make create-redis`.

## Deploying the Go Server

To deploy our custom Go Server (once the image has been built and push to
[Google Container Registry](https://cloud.google.com/container-registry/), run `make create-server`.
