#!/bin/sh
set -u
set -e
set -x

rm -rf internal go.mod go.sum main.go registry.go openapi.yml

sqlc-http -m booktest -tracing -metric
