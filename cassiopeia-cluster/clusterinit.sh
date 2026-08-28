#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

kubectl apply -f "$SCRIPT_DIR/infra/namespace.yaml"
kubectl apply -f "$SCRIPT_DIR/infra/traefik.yaml"
kubectl apply -f "$SCRIPT_DIR/infra/gateway.yaml"
