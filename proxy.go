package main

var proxy_template = `---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  annotations:
    getambassador.io/config: |
      ---
      apiVersion: ambassador/v0
      kind: Mapping
      name: map-{{.Name}}
      grpc: true
      prefix: /helloworld.Greeter/
      rewrite: /helloworld.Greeter/
      service: {{.Name}}:{{.Port}}
spec:
  type: NodePort
  ports:
  - nodePort: {{.NodePort}}
    port: {{.Port}}
    targetPort: {{.TargetPort}}
  selector:
    app: {{.Name}}
`
