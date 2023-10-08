#!/bin/sh

set -ex

docker build -t timoni/metrics-grafana .
docker push timoni/metrics-grafana
