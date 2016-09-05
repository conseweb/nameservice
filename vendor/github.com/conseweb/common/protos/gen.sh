#!/usr/bin/env bash

WORKDIR="./"

protoc -I ./ *.proto --go_out=plugins=grpc:.