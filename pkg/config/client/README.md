## Using command line client to generate yml ##

```
$ go build github.com/Peripli/istio-broker-proxy/pkg/config/client
$ ./client --endpoint <> --port <> --service <> --virtual-service <>
$ ./client --port <> --service <> --virtual-service <>
```

e.g.
```
$ ./client -client -port 5555 -service pinger -virtual-service pinger.istio.cf.dev01.aws.istio.sapcloud.io -system-domain istio.cf.dev01.aws.istio.sapcloud.io
```
