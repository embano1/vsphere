#!/usr/bin/env bash

: "${KO_DOCKER_REPO:?"You must set environment variable 'KO_DOCKER_REPO'"}"

export GO111MODULE=on

cat << EOF | ko resolve --platform=linux/amd64 -Bf -
images:
- ko://github.com/embano1/vsphere/test/images/client
EOF
