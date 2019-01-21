#!/bin/sh
docker build . -t $HUB/client
docker push $HUB/client
