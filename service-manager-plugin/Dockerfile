FROM  gcr.io/sap-se-gcp-istio-dev/sb-proxy-k8s

WORKDIR /app

RUN apk add --no-cache bash gawk sed grep bc coreutils

ADD service-manager-plugin/service-manager-plugin.so /app/service-manager-plugin.so

ENTRYPOINT [ "./main"]