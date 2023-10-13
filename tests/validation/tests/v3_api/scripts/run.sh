#!/bin/bash

IMAGE_NAME="rancher-validation-networkchecks-provisioning-network-checks_1696949864449_test"
docker run --env-file .env ${IMAGE_NAME} pytest -k "test_wl or test_connectivity or test_ingress or test_service_discovery or test_websocket" -v -s tests/v3_api/