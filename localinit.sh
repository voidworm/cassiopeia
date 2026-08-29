#!/usr/bin/env bash
set -euo pipefail

#setup plain infra
docker compose down -v
docker compose build
docker compose up -d
docker compose run --rm cassiopeia-backend seed initial
docker compose run --rm cassiopeia-backend seed personal
