#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

#setup plain infra
#init namespaces
kubectl apply -f "$SCRIPT_DIR/infra/namespace.yaml"
#setup traefik root infra
kubectl apply -f "$SCRIPT_DIR/infra/traefik.yaml"
#setup gatewayclass & gateway
kubectl apply -f "$SCRIPT_DIR/infra/gateway.yaml"

#setup argo
#apply argo root deployment
kubectl apply --server-side -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml -n argocd
#patch config map to allow insecure comm due to lacking TLS certs, and make
#argo aware it's served under the /argocd path prefix (not domain root)
kubectl -n argocd patch configmap argocd-cmd-params-cm --type=merge \
  -p='{"data":{"server.insecure":"true","server.basehref":"/argocd","server.rootpath":"/argocd"}}'
#restart the deployment to load updated configmap
kubectl -n argocd rollout restart deployment argocd-server
#apply httproute to expose argo ui at cassiopeia.local/argocd
kubectl apply -f "$SCRIPT_DIR/infra/argocd-route.yaml"

#apply application for cassiopeia front/backend
kubectl apply -f "$SCRIPT_DIR/argocd/backend-application.yaml"
kubectl apply -f "$SCRIPT_DIR/argocd/frontend-application.yaml"
