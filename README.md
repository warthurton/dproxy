# dproxy

## Config

```yaml
bind_addr: 0.0.0.0:8000
direktiv_addr: 192.168.139.128
routes:
  - namespace: test
    token: "blah"
  - alias: test2
    namespace: test
    token: "blah2"
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

The following additional optional query parameters can be passed along as-is: `ctype`, `field`, `raw-output`.
