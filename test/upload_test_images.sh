#!/usr/bin/env bash

: "${KO_DOCKER_REPO:?"You must set environment variable 'KO_DOCKER_REPO'"}"

export GO111MODULE=on

cat << EOF | ko resolve -Bf -
images:
- ko://github.com/embano1/vsphere/test/images/client
EOF
