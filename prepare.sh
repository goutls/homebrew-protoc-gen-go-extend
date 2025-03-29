#!/bin/zsh

# make directory were store files
rm -Rf temp || true
mkdir temp || true

repo=github.com/goutls/protoc-gen-go-extend
releaseViewFileName=temp/releases.json


gh release list --json createdAt,isDraft,isLatest,isPrerelease,name,publishedAt,tagName -R ${repo} > ${releaseViewFileName}
go run cmd/release_view/ ${releaseViewFileName}