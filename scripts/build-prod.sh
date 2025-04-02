#!/bin/bash

# make sure binary exists
if [ ! -f ./bin/${PROJECT_NAME} ]; then
  echo "Binary not found. Please run 'make build' first."
  exit 1
fi

# make sure version is passed from arguments
if [ -z "$1" ]; then
    echo "Please provide a version number as an argument."
    exit 1
fi
if [[ ! "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-alpha)?$ ]]; then
    echo "Invalid version number. Please use the format v0.0.0 or v0.0.0-alpha."
    exit 1
fi

TAG=$1

docker build \
    -t ${GAR_REGION}-docker.pkg.dev/${GAR_PROJECT_ID}/${GAR_REPOSITORY}/${PROJECT_NAME}:${TAG} \
    -t ${GAR_REGION}-docker.pkg.dev/${GAR_PROJECT_ID}/${GAR_REPOSITORY}/${PROJECT_NAME}:latest \
    -f Dockerfile \
    .