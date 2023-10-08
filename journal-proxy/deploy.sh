#!/bin/bash

CLUSTER=${CLUSTER:-supervisor}
KUBECONFIG=$(k3d kubeconfig write $CLUSTER)

pwd=$(pwd)
cd $(dirname $(realpath $0))


commitSHA=$(kubectl --kubeconfig $KUBECONFIG get deploy -n timoni-journal journal-proxy -o jsonpath='{.spec.template.spec.containers[].image}' | awk -F: '{print $2}')
if [ -z "$commitSHA" ]; then
    commitSHA=$(git rev-parse HEAD)
fi

IMAGE_TAG="cs-local.syslabit.org/journal-proxy:$commitSHA"

set -ex

# Build
CGO_ENABLED=0 go build .
docker build -t $IMAGE_TAG .

# Add to k3d
k3d image import $IMAGE_TAG -c $CLUSTER 

# Restart
kubectl --kubeconfig $KUBECONFIG rollout restart -n timoni-journal deployment/journal-proxy

cd $pwd
