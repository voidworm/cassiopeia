#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-cassiopeia}"

echo "Tearing down cassiopeia-backend + postgres in namespace '${NAMESPACE}'..."

kubectl delete deployment cassiopeia-backend cassiopeia-postgres -n "$NAMESPACE" --ignore-not-found
kubectl delete pvc cassiopeia-postgres-data -n "$NAMESPACE" --ignore-not-found

echo "Backend Deployment and PVC deleted."
