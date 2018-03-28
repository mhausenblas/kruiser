# Kruiser

Note: tested on Minikube with v1.8; also, this is an experiment ;)

## Preparation

Create namespace:

```bash
$ kubectl create namespace kruiser
```

## gRPC service with NGINX sidecar

The gRPC test service is [mhausenblas/yages](https://github.com/mhausenblas/yages), using [NGINX 1.13.10](https://www.nginx.com/blog/nginx-1-13-10-grpc/) as a sidecar. 

Deploy gRPC service + sidecar proxy with:

```
$ kubectl -n kruiser apply -f kruiser.yaml
```

Clean up with:

```
$ kubectl -n kruiser delete all -l=app=kruiser
```

## Invoke from jump pod

Launch a gRPC enabled jump pod:

```
$ kubectl -n kruiser run -it --rm gumpod --restart=Never --image=quay.io/mhausenblas/gump:0.1
```

Directly accessing the gRPC service:

```
/go $ grpcurl --plaintext kruiser:9000 yages.Echo.Ping
```

Accessing the gRPC service via NGINX proxy:

```
/go $ grpcurl --plaintext kruiser:8080 yages.Echo.Ping
```

## Manual set up of NGINX

To configure NGINX manually, exec into the pod (assuming below here that the `kruiser` pod is `kruiser-856686799d-j792v`) and using container `proxy`:

```
$ kubectl -n kruiser exec -it -c proxy kruiser-856686799d-j792v -- sh
```

In the `proxy` container, for example you can do:

```bash
$ cat << EOF > /etc/nginx/conf.d/grpc-proxy.conf
server {
    listen 8080 http2;

    access_log  /var/log/nginx/grpc.log;

    location / {
        grpc_pass grpc://localhost:9000;
    }
}
EOF
```

Re-start NGINX with `nginx -s reload` to apply changes.

See also logs via: `cat /var/log/nginx/grpc.log`.