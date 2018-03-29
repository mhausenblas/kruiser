# Kruiser

A proxy that transparently exposes gRPC Kubernetes services cluster-externally.

Using [NGINX 1.13.10](https://www.nginx.com/blog/nginx-1-13-10-grpc/) as a proxy, `kruiser` 
watches services labelled with `grpc=expose` and proxies them to the public using a service of type NodePort on port `32123`.

![architecture](img/kruiser-arch.png)

Note: so far only tested on Minikube v0.24 with Kubernetes v1.8.



## Use cases

There are two

### UC1: inter-cluster within the enterprise

### UC2: public services

## Usage

Either automatic or static

## Prerequisites 

First, clone this repository with `git clone https://github.com/mhausenblas/kruiser.git && cd kruiser`.

It's considered a good practice to create namespaces for related apps rather than dumping all into the `default` namespace.
And indeed, the usage instructions throughout assume you've created a namespace `kruiser`, for example, like so:

```bash
$ kubectl create namespace kruiser
```

The gRPC demo service used throughout here is a simple echo service: [mhausenblas/yages](https://github.com/mhausenblas/yages). 
As a generic gRPC client we use [fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) here 
which you can either install locally (if you have Go) or as a container using [quay.io/mhausenblas/gump:0.1](https://quay.io/repository/mhausenblas/gump?tag=0.1&tab=tags).

## gRPC service and standalone NGINX proxy

You can use [static/yages.yaml](static/yages.yaml) and [static/kruiser.yaml](static/kruiser.yaml) as a static boilerplate as described in the following.

### Setup

Deploy the demo gRPC service with:

```bash
$ kubectl -n kruiser apply -f static/ping.yaml
```

Deploy the NGINX proxy with:

```
$ kubectl -n kruiser apply -f static/kruiser.yaml
```

When done, clean up with:

```
$ kubectl -n kruiser delete all -l=app=yages
$ kubectl -n kruiser delete all,cm -l=app=kruiser
```

### Invoke

Option 1: Launch a gRPC enabled jump pod and access the gRPC service via the NGINX proxy from within the cluster:

```bash
$ kubectl -n kruiser run -it --rm gumpod --restart=Never --image=quay.io/mhausenblas/gump:0.1

/go $ grpcurl --plaintext kruiser:8080 yages.Echo.Ping
```

Option 2: Assuming you're using Minikube and you have `grpcurl` installed locally, you can access the gRPC service from outside the cluster as shown here:

```bash
$ grpcurl --plaintext $(minikube ip):32123 yages.Echo.Ping
```

## gRPC service with NGINX proxy sidecar

You can use [static/sidecar-kruise.yaml](static/sidecar-kruise.yaml) as a static boilerplate as described in the following.

Note that in addition to NGINX being deployed as a sidecar, the config is also different. Here, `/etc/nginx/nginx.conf` is directly overridden, in contrast to the standalone case where a new config file `/etc/nginx/conf.d/grpc-proxy.conf` is added. 

### Setup

Deploy gRPC service + sidecar proxy with:

```bash
$ kubectl -n kruiser apply -f static/sidecar-kruiser.yaml
```

Clean up with:

```bash
$ kubectl -n kruiser delete all,cm -l=app=kruiser
```

### Invoke

Launch a gRPC enabled jump pod:

```bash
$ kubectl -n kruiser run -it --rm gumpod --restart=Never --image=quay.io/mhausenblas/gump:0.1
```

Now you can directly access the gRPC service from within the cluster:

```bash
/go $ grpcurl --plaintext kruiser:9000 yages.Echo.Ping
```

Accessing the gRPC service via NGINX proxy from within the cluster looks like this:

```bash
/go $ grpcurl --plaintext kruiser:8080 yages.Echo.Ping
```

## Manual set up of NGINX

To toy around and try new things, configure NGINX manually. For that exec into the pod (assuming below here that the `kruiser` pod is `kruiser-856686799d-j792v`) and use container `proxy`:

```bash
$ kubectl -n kruiser exec -it -c proxy kruiser-856686799d-j792v -- sh
```

In the `proxy` container, for example you can do:

```bash
$ cat << EOF > /etc/nginx/conf.d/grpc-proxy.conf
server {
    listen 8080 http2;

    access_log  /tmp/grpc.log;

    location / {
        grpc_pass grpc://localhost:9000;
    }
}
EOF
```

To apply the config and re-start NGINX do: `nginx -s reload`. See also the logs at: `cat /tmp/grpc.log`.