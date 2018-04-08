# Kruiser

A proxy that transparently exposes gRPC Kubernetes services cluster-externally.

Using [Ambassador](https://www.getambassador.io/) as gRPC proxy, `kruiser` 
watches deployment in a target namespace that are labelled with `grpc=expose`. When it finds such a deployment, it creates a corresponding service—of type `NodePort` on ports from `31000` upwards—that proxies traffic to its pods from outside the cluster.

So far, I've tested `kruiser` on Minikube v0.24 with Kubernetes v1.8 and v1.9, as well as on GKE with Kubernetes v1.9 with and without RBAC.

- [Use cases](#use-cases)
  - [UC1: inter-cluster within the enterprise](#uc1-inter-cluster-within-the-enterprise)
  - [UC2: public services](#uc2-public-services)
- [Install](#install)
- [Use](#use)
  - [Example gRPC demo services](#example-grpc-demo-services)
  - [Walkthroughs](#walkthroughs)
    - [Minikube](#minikube)
    - [GKE](#gke)
    - [Cleanup](#cleanup)

## Use cases

There are two main use cases:

### UC1: inter-cluster within the enterprise

Imagine two or more clusters deployed within, say, a data center in an enterprise. In order for gRPC services to communicate across clusters, you need to proxy the traffic from one cluster to another.

### UC2: public services

If you want to make your gRPC service publicly available, you need to somehow expose it, routing traffic from outside the cluster to the cluster-internal service.


## Install 

First, clone this repository with `git clone https://github.com/mhausenblas/kruiser.git && cd kruiser`.

Creating a namespaces for related apps rather than dumping all into the `default` namespace is a good practice, so let's do that first:

```bash
$ kubectl create namespace kruiser
```

## Use

### Example gRPC demo services

The two example gRPC [demo services/](demo-services/) used below here are:

- A simple echo service [yages.Echo](https://github.com/mhausenblas/yages/blob/master/main.go) available via `quay.io/mhausenblas/yages:0.1.0`
- The reference [helloworld.Greeter](https://github.com/grpc/grpc-go/blob/master/examples/helloworld/greeter_server/main.go) available via `quay.io/mhausenblas/grpc-gs:0.2`

As a generic gRPC client we use [fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) which you can either install locally, if you have Go installed, or as a container via the [quay.io/mhausenblas/gump:0.1](https://quay.io/repository/mhausenblas/gump?tag=0.1&tab=tags) container image.


### Walkthroughs

In the following, I'll walk you through how you can use `kruiser` in a static manner, that is, manually exposing gRPC services cluster-externally. Along the way I explain how `kruiser` works.


```bash
$ kubectl create namespace kruiser
```

#### Minikube 

First, install Ambassador with:

```bash
$ kubectl -n kruiser apply -f ambassador/admin.yaml
```

Next, deploy the two gRPC demo services:

```bash
$ kubectl -n kruiser apply -f demo-services/
```

Now you can invoke each of the gRPC demo services from outside Minikube like so:

```bash
$ grpcurl --plaintext $(minikube ip):31001 yages.Echo.Ping

$ grpcurl --plaintext  -d '{ "name" : "Michael" }' $(minikube ip):31000 helloworld.Greeter.SayHello
```

Alternatively, you can access one of the gRPC services via the gRPC jump pod like so:

```bash
$ kubectl -n kruiser run -it --rm gumpod \
          --restart=Never --image=quay.io/mhausenblas/gump:0.1

/go $ grpcurl --plaintext ping:9000 yages.Echo.Ping
```

#### GKE

Note that the GKE deployment in the following uses RBAC for access control.

As a preparation, you need to give your user certain rights.

```bash
$ cat ambassador/gke-crb.yaml | \
  sed s/__USER__/$(gcloud projects get-iam-policy $(gcloud config get-value core/project) | grep -m 1 user | awk '{split($0,u,":"); print u[2]}')/g | \
  kubectl -n kruiser apply -f -
```

Above, we replace the `__USER__` placeholder in `ambassador/gke-crb.yaml` with the value of the user name of the active GKE project before creating the respective cluster-role binding.

Next, install Ambassador with:

```bash
$ kubectl -n kruiser apply -f ambassador/admin-rbac.yaml
```

And now deploy the two gRPC demo services:

```bash
$ kubectl -n kruiser apply -f demo-services/
```

To be able to access the services from outside the GKE cluster we first have to find values for external IPs of cluster nodes (store them for example in an env var `NODE_IP`):

```bash
$ kubectl get nodes --selector=kubernetes.io/role!=master \ 
                    -o jsonpath={.items[*].status.addresses[?\(@.type==\"ExternalIP\"\)].address}
```

Now, finally, you can invoke each of the gRPC demo services from outside the GKE cluster like so: 

```bash
$ grpcurl --plaintext $(NODE_IP):31001 yages.Echo.Ping

$ grpcurl --plaintext  -d '{ "name" : "Michael" }' $(NODE_IP):31000 helloworld.Greeter.SayHello
```

#### Cleanup

When done, clean up with:

```bash
$ kubectl delete ns kruiser
```
