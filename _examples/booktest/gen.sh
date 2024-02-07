#!/bin/sh
set -u
set -e
set -x

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

rm -rf internal go.mod go.sum main.go registry.go openapi.yml

sqlc generate
sqlc-http -m booktest -tracing -metric
