#!/bin/bash

ENV=$1
NAMESPACE=$2

kubectl create configmap secondbaser-config-map --from-file=config/conf/$ENV.json -n $NAMESPACE