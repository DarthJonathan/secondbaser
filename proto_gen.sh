#!/bin/bash

PROTO_PATH="./api/proto"
GEN_PATH="./api/go_gen"

protoc  --proto_path=${PROTO_PATH} \
        --go_out=${GEN_PATH} --go_opt=paths=source_relative \
        --go-grpc_opt=paths=source_relative \
        --go-grpc_opt=require_unimplemented_servers=false \
        --go-grpc_out=${GEN_PATH} \
        TransactionalService.proto