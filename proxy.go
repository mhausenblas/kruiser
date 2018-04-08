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
      prefix: /{{.FQServiceName}}/
      rewrite: /{{.FQServiceName}}/
      service: {{.Name}}:{{.Port}}
spec:
  type: NodePort
  ports:
  - port: {{.Port}}
    targetPort: {{.Port}}
  selector:
    app: {{.Name}}
`
