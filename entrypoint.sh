#!/bin/bash

DEPLOY_ENV=$1

exec ./secondbaser -config-file=./config/conf/${DEPLOY_ENV}.json