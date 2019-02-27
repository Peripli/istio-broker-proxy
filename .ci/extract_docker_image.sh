#!/bin/bash -eu

OUTPUT=$(realpath $1)

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

ISTIO_VERSION=$(awk '/version:/ { print($2); }' "$SCRIPTDIR/../istio/Chart.yaml" )

curl -L -O "https://github.com/istio/istio/releases/download/$ISTIO_VERSION/istio-$ISTIO_VERSION-linux.tar.gz" 
tar xzf "istio-$ISTIO_VERSION-linux.tar.gz"

pushd "istio-$ISTIO_VERSION"

helm init --client-only 

if [ ! -d ./install/kubernetes/helm/istio/charts ] ; then
  echo "Detected newer version (1.1+) of istio, running helm dep update" >&2
  helm repo add istio.io https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/release-1.1-latest-daily/charts/ 
  helm dep update "./install/kubernetes/helm/istio"
fi

helm template \
  --namespace=istio-system \
  --values "$SCRIPTDIR/../docker/values.yaml" \
  --set global.mtls.enabled=false \
  --set global.controlPlaneSecurityEnabled=false \
  install/kubernetes/helm/istio >> install/kubernetes/istio.yaml


awk '/ image: "/ { print($0) }' install/kubernetes/istio.yaml | awk '-F"' ' {print($2)}' | sort -u > $OUTPUT

popd

rm -rf "istio-$ISTIO_VERSION-linux.tar.gz" "istio-$ISTIO_VERSION"
