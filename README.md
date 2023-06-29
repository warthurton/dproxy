# dproxy

## Config

```yaml
bind_addr: 0.0.0.0:8080
direktiv_addr: equinix.direktiv.io
insecure_skip_verify: true
routes:
  - alias: html
    namespace: html-test
    token: "v4.local.token"
  - alias: json
    namespace: json-test
    token: "v4.local.token"
```

## Proxying

Using the above config, incoming requests to:
```
http://localhost:8000/n/test/w/hello.yaml
```

Are proxied to:
```
https://192.168.139.128/api/namespaces/alan/tree/hello.yaml?op=wait
```

## Installation

Install this using the following information. 

1. Create the following or modify kubernetes/install.yaml and add the following information:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dproxy-cfg-cm
data:
  config.yaml: |
      bind_addr: 0.0.0.0:8080
      direktiv_addr: prod.direktiv.io
      routes:
        - alias: html
          namespace: html-test
          token: "v4.local.token"
        - alias: json
          namespace: json-test
          token: "v4.local.token"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dproxy
  labels:
    app: dproxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dproxy
  template:
    metadata:
      annotations:
        linkerd.io/inject: disabled
      labels:
        app: dproxy
    spec:
      volumes:
      - name: dproxyconf
        configMap:
          name: dproxy-cfg-cm
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532        
        runAsGroup: 65532
      containers:
        - name: dproxy
          image: gcr.io/dproxy:1.0
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
          volumeMounts:
          - name: dproxyconf
            mountPath: "/config"
            readOnly: false
---
apiVersion: v1 
kind: Service
metadata:
  name: dproxy-service
spec:
  selector:
    app: dproxy
  ports:
    - port: 8080
---
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
  name: dproxy-route
spec:
  http:
  - name: dproxy-receiver
    match:
      hosts:
      - prod.direktiv.io
      paths:
      - "/dproxy/*"
    backends:
    - serviceName: dproxy-service
      servicePort: 8080
```

2. Install the service on the Kubernetes platform using the file create or the `kuberetes/install.yaml`:

```sh
kubectl apply -f kubernetes/install.yaml
```

3. The following additional optional query parameters can be passed along as-is: `ctype`, `field`, `raw-output`.

As an example, here is a workflow that produce HTML output:

```yaml
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform: 'jq({ result: ("<!DOCTYPE html><html><body><h1>My First Heading</h1><p>My first paragraph.</p></body></html>" | @base64)})'
```

Executing this directly from a browser:

https://prod.direktiv.io/dproxy/n/html/w/return-html?field=result&raw-output=true 