# Kruiser

A proxy that transparently exposes gRPC Kubernetes services cluster-externally.

Using back-ends such as [NGINX 1.13.10](https://www.nginx.com/blog/nginx-1-13-10-grpc/) or [Ambassador](https://www.getambassador.io/) as proxies, `kruiser` 
watches services labelled with `grpc=expose` and proxies them to the public using a service of type NodePort on port `32123`.

![architecture](img/kruiser-arch.png)

Note: so far tested on Minikube v0.24 and v0.25 with Kubernetes v1.8 and v1.9.

- [Use cases](#use-cases)
    - [UC1: inter-cluster within the enterprise](#uc1-inter-cluster-within-the-enterprise)
    - [UC2: public services](#uc2-public-services)
- [Install](#install)
- [Use](#use)

## Use cases

There are two

### UC1: inter-cluster within the enterprise

### UC2: public services

## Install 

First, clone this repository with `git clone https://github.com/mhausenblas/kruiser.git && cd kruiser`.

It's considered a good practice to create namespaces for related apps rather than dumping all into the `default` namespace.
And indeed, the usage instructions throughout assume you've created a namespace `kruiser`, for example, like so:

```bash
$ kubectl create namespace kruiser
```

The gRPC demo service used throughout here is a simple echo service: [mhausenblas/yages](https://github.com/mhausenblas/yages). 
As a generic gRPC client we use [fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) here 
which you can either install locally (if you have Go) or as a container using [quay.io/mhausenblas/gump:0.1](https://quay.io/repository/mhausenblas/gump?tag=0.1&tab=tags).

If you're interested in how `kruiser` works or want to extend it, check out the [how does it work](how-does-it-work.md) notes.

## Use

Deploy the demo gRPC services with:

```bash
$ kubectl -n kruiser apply -f static/ping.yaml,static/gs.yaml
```

Deploy the NGINX proxy with:

```bash
$ kubectl -n kruiser apply -f static/kruiser.yaml
```

Invoke it like so, for example:

```bash
$ grpcurl --plaintext kruiser:8080 yages.Echo.Ping
```

When done, clean up with:

```
$ kubectl -n kruiser delete all -l=app=yages
$ kubectl -n kruiser delete all,cm -l=app=kruiser
```
