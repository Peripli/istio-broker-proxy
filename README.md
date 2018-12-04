# Install pre-commit hook

The hook will
* Call `go fmt`

```
cd .git/hooks
ln -s ../../hooks/pre-commit pre-commit
```

# istio-broker

Forward all requests to the service fabrik.

## Steps to deploy

* Deploy the app:
```
  cf push istio-broker
```

* Delete the service fabrik broker (There might be some services, that have to be deleted before).
```
  cf delete-service-broker service-fabrik-broker
```

* Create a new service broker with the credentials of the service fabrik and the URL of the pushed app. The credentials are found in deployments/service-fabrik/credentials.yml as credentials.broker.user and credentials.broker.password.
```
cf create-service-broker istio-broker <user> <password> https://istio-broker.cfapps.dev01.aws.istio.sapcloud.io
```

* Create an application security group using [sec_group.json](sec_group.json)
```
cf create-security-group istio-broker-service-fabrik sec_group.json
```

* List services using `cf service-access` and enable services using `cf enable-service-access`.

* Check that services are available
```
cf marketplace
```


## Steps to validate

* Use service broker
```
cf service-brokers
cf create-service postgresql v9.4-dev mydb
cf delete-service mydb
```

* Check tracking
```
curl https://istio-broker.cfapps.dev01.aws.istio.sapcloud.io/info
```
