#!/bin/sh
set -u
set -e
set -x

rm -rf internal proto api go.mod go.sum *.go openapi.yml

sqlc-http -m authors -migration-path sql/migrations -litestream
