#!/bin/sh
set -u
set -e
set -x

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

rm -rf internal proto api go.mod go.sum *.go openapi.yml

sqlc generate
sqlc-http -m authors -migration-path sql/migrations -litefs -litestream
