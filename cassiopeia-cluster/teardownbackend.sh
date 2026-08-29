#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-cassiopeia}"

echo "Tearing down cassiopeia-backend + postgres in namespace '${NAMESPACE}'..."

kubectl delete deployment cassiopeia-backend cassiopeia-postgres -n "$NAMESPACE" --ignore-not-found
kubectl delete service cassiopeia-backend cassiopeia-postgres -n "$NAMESPACE" --ignore-not-found
kubectl delete pvc cassiopeia-postgres-data -n "$NAMESPACE" --ignore-not-found

echo "Done. Argo CD selfHeal will recreate the Deployments/Services from the chart; postgres will come back up empty and re-run Migrate()/PreseedIfEmpty()."
