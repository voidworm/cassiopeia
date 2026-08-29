#!/usr/bin/env bash
set -euo pipefail

#setup plain infra
docker compose down -v
docker compose up -d
docker compose build
docker compose run --rm cassiopeia-backend seed initial
docker compose run --rm cassiopeia-backend seed personal
