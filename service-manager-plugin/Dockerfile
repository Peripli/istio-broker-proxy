FROM  gcr.io/sap-se-gcp-istio-dev/sb-proxy-k8s

WORKDIR /app

ADD service-manager-plugin/service-manager-plugin.so /app/service-manager-plugin.so

ENTRYPOINT [ "./main", "--plugin" , "service-manager-plugin.so" ]