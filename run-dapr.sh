#! /usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

dapr run \
  --app-id ucp \
  --app-port 8081 \
  --config "$SCRIPT_DIR/components/config.yaml" \
  --resources-path "$SCRIPT_DIR/components" \
  --dapr-grpc-port 50001