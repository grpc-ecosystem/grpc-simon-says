# Simon Server

This is the game server that runs behind the multiplayer version of [Simon](https://en.wikipedia.org/wiki/Simon_\(game\))!

It runs on several pieces of technology:
- [Go](https://golang.org/)
- [gRPC](http://www.grpc.io/)
- [Kubernetes](http://kubernetes.io/)
- [Redis](http://redis.io/)

## TL;DR

To deploy the entire thing, from creating the Kubernetes cluster, to building the docker images and having them run
on the cluster, use `make all`, which will run all the steps in order, and give you a running server.

## A More Detailed Description

A break down of each of the pieces in this server.

### Kubernetes Cluster

To create the Kubernetes cluster, read the [Infrastructure ReadMe](./infrastructure/README.md).

### Building the Docker Images

To generate the gRPC code, compile the go code into a static binary, and create the docker file, 
run `make build`.

To push this Docker image up to the [Google Container Registry](https://cloud.google.com/container-registry/), run `make push`.

### Deploying to Kubernetes

To deploy the image, and the Redis service to Kubernetes, read the [Deploy ReadMe](./deploy/README.md)

## Licence
Apache 2.0

This is not an official Google Product.
