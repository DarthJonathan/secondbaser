#!/bin/bash

BUILD=$1
REGISTRY=asia.gcr.io
REPO=trakkie-id
APP=secondbaser
TAG=trakkie/secondbaser

docker build \
        -t $TAG .
docker tag $TAG $REGISTRY/$REPO/$APP:$BUILD
exec docker push $REGISTRY/$REPO/$APP:$BUILD
