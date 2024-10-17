#!/usr/bin/env bash
# -*- coding: utf-8 -*-

set -ex

registry="${REGISTRY:-docker.io}"

for dockerfile in base/*.Dockerfile; do
	docker build -t "$registry/zeabur/zbpack-$(basename "$dockerfile" .Dockerfile):latest" -f "$dockerfile" .
	docker push "$registry/zeabur/zbpack-$(basename "$dockerfile" .Dockerfile):latest"
done
