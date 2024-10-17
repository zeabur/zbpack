FROM docker.io/library/golang:1.23-alpine as builder
RUN apk add --no-cache build-base cmake
