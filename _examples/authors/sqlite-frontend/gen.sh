#!/bin/sh
set -u
set -e
set -x

#go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

rm -rf internal view go.mod go.sum *.go openapi.yml

sqlc generate
sqlc-http -m sqlite-htmx -migration-path sql/migrations -frontend
