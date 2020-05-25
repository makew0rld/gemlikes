#!/usr/bin/env bash

set -e

mkdir -p build

cd add-comment
go build
mv add-comment ../build

cd ../like
go build
mv like ../build

cd ../view
go build
mv view ../build

cd ..