#!/bin/bash

TRIM_JOB_NAME=$(basename "$JOB_NAME")
IMAGE_TAG=rancher-validation-"${TRIM_JOB_NAME}""${BUILD_NUMBER}"
IMAGE_NAME="${IMAGE_NAME:-$IMAGE_TAG}"

docker rm -f {TEST_CONTAINER}
docker rmi ${IMAGE_NAME}