#!/bin/sh
set -u
set -e
set -x

rm -rf internal view go.mod go.sum *.go openapi.yml

sqlc-http -m sqlite-htmx -migration-path sql/migrations -frontend
