# How does it work?

In a nutshell, `kruiser` uses [NGINX 1.13.10](https://www.nginx.com/blog/nginx-1-13-10-grpc/) as a proxy for a Kubernetes services running a gRPC service.

In the following, I'll walk you through how you can use the templates in [static/](static/) to set up `kruiser` in a static manner, that is, manually exposing gRPC services cluster-externally. Along the way I explain how `kruiser` works.

- [Preparation](#preparation)
- [gRPC service and standalone NGINX proxy](#grpc-service-and-standalone-nginx-proxy)
    - [Setup](#setup)
    - [Invoke](#invoke)
- [gRPC service with NGINX proxy sidecar](#grpc-service-with-nginx-proxy-sidecar)
    - [Setup](#setup-1)
    - [Invoke](#invoke-1)
- [Manual set up of NGINX](#manual-set-up-of-nginx)


## Preparation

Create a namespace such as `kruiser`, for example, like so:

```bash
$ kubectl create namespace kruiser
```

The gRPC demo service used throughout here is a simple echo service: [mhausenblas/yages](https://github.com/mhausenblas/yages). 
As a generic gRPC client we use [fullstorydev/grpcurl](https://github.com/fullstorydev/grpcurl) here 
which you can either install locally (if you have Go installed) or as a container using [quay.io/mhausenblas/gump:0.1](https://quay.io/repository/mhausenblas/gump?tag=0.1&tab=tags). Alternatively, you can use the [YAGES client](client/) (if you gave Go installed).

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

The core of the `kruiser` deployment is the following `ConfigMap`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginxconf
  labels:
    app: kruiser
data:
  config: | 
    server {
      listen 8080 http2;
      
      access_log /dev/stdout;
      error_log /dev/stderr warn;

      location / {
        grpc_pass grpc://ping:9000;
      }
    } 
```

From above `ConfigMap` you can see that it configures NGINX to passthrough HTTP/2 traffic from `localhost:8080` to `grpc://ping:9000`.

### Invoke

Option 1: Launch a gRPC enabled jump pod and access the gRPC service via the NGINX proxy from within the cluster:

```bash
$ kubectl -n kruiser run -it --rm gumpod --restart=Never --image=quay.io/mhausenblas/gump:0.1

/go $ grpcurl --plaintext kruiser:8080 yages.Echo.Ping
```

Option 2: Assuming you're using Minikube and you have `grpcurl` installed locally (requires Go), you can access the gRPC service from outside the cluster as shown here:

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