
# Kubernetes client problems

### Plain kubernetes client


```
clientset, err := kubernetes.NewForConfig(config)
```

### Rest client

#### Variant 1

```
clientset, err := kubernetes.NewForConfig(config)
clientset.RESTClient().Namespace(out.GetObjectMeta().Namespace).
                      		Resource(crd.ResourceName(schema.Plural)).
                      		Body(out).
                      		Do().
                      		Get()
```

**No serializer found for**

The simple RESTClient has only decoders installed.

#### Variant 2


```
config.GroupVersion = &schema.GroupVersion{
    Group:   "config.istio.io",
    Version: "v1alpha2",
}
config.APIPath = "/apis"
config.ContentType = runtime.ContentTypeJSON

types := runtime.NewScheme()
schemeBuilder := runtime.NewSchemeBuilder(
    func(scheme *runtime.Scheme) error {
        metav1.AddToGroupVersion(scheme, *config.GroupVersion)
        return nil
    })
err := schemeBuilder.AddToScheme(types)
config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(types)}

client, err := rest.RESTClientFor(config)

client.Post().
		Namespace(out.GetObjectMeta().Namespace).
		Resource(crd.ResourceName(schema.Plural)).
		Body(out).
		Do().
		Get()

```

**the server could not find the requested resource (post serviceentries.config.istio.io)**