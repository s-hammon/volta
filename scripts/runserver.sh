#!/bin/bash

docker run --rm \
    --name srv-volta \
    --network host \
    -v ${HOME}/.config/gcloud:/root/.config/gcloud \
    volta:latest serve -d ${DATABASE_URL}