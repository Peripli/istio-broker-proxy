# istio-broker

Forward all requests to the service fabrik.

* Deploy:
```
  cf push istio-broker
```

* Delete the service fabrik broker (Threre might be some services, that have to be deleted before).
* Create a new service broker with the credentials of the service fabrik and the URL of the pushed app.
* Create an application security group using [sec_group.json](sec_group.json)
* List services using `cf service-access` and enable services using `cf enable-service-access`.
* Create services
