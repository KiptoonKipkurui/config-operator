#!/bin/bash
set -euo pipefail

TYPE=${1:-validate}


if [ "${TYPE}" != "mutate" ] && [ "${TYPE}" != "validate" ]; then
  echo "invalid webhook type: ${TYPE}"
  echo "usage: $0 [mutate|validate]"
  exit 1
fi;
SCRIPT_DIR=$(cd $(dirname $0);pwd)

curl -k --request POST "https://127.0.0.1:9443/validate-config-configop-com-v1alpha1-configop" \
        -H "Content-Type: application/json" \
        --data-binary @"${SCRIPT_DIR}/admission-review-request.json"