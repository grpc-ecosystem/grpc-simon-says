# Create Kubernetes Infrastructure

This manages creating the [Kubernetes](http://kubernetes.io/) cluster on
[Google Container Engine](https://cloud.google.com/container-engine/)
via [Deployment Manager](https://cloud.google.com/deployment-manager/)

## Creating the Cluster
To create the cluster, run `make deploy`

## Authenticating kubectl
Once you have a cluster up and running, you will need to get the credentials
for Kubernetes, so that [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/) can be used to 
interact with the Kubernetes cluster.

To do that, run `make auth` and the credentials will be set up for you.
