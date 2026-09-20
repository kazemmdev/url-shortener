#!/usr/bin/env bash
# Generates secrets.yaml (gitignored) next to this script. Idempotent: if a
# url-shortener-secrets Secret already exists in the cluster, its password is
# reused so re-running `make k8s-up` doesn't drift the password out from
# under an already-initialized SQL Server (which bakes the sa password into
# its data files at first boot and ignores MSSQL_SA_PASSWORD after that —
# a changed Secret plus any later container restart was enough to break
# login on a currently-running instance).
#
# To rotate the password on purpose: ROTATE=1 ./create-secrets.sh, then wipe
# sqlserver's data too (`make k8s-down` or delete its PVC) so it
# re-initializes with the new password — otherwise it keeps the old one.
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

kubectl create namespace url-shortener --dry-run=client -o yaml | kubectl apply -f - >/dev/null

EXISTING_PASSWORD=""
if [[ "${ROTATE:-0}" != "1" ]]; then
  EXISTING_PASSWORD="$(kubectl -n url-shortener get secret url-shortener-secrets \
    -o jsonpath='{.data.MSSQL_SA_PASSWORD}' 2>/dev/null | base64 -d 2>/dev/null || true)"
fi

if [[ -n "${EXISTING_PASSWORD}" ]]; then
  PASSWORD="${EXISTING_PASSWORD}"
  echo "Reusing existing MSSQL_SA_PASSWORD from the cluster (set ROTATE=1 to generate a new one)."
else
  PASSWORD="${MSSQL_SA_PASSWORD:-$(openssl rand -base64 24 | LC_ALL=C tr -dc 'A-Za-z0-9')Aa1!}"
  if [[ "${ROTATE:-0}" == "1" ]]; then
    echo "ROTATE=1: generating a new password. Run 'make k8s-down' (or delete the" >&2
    echo "sqlserver PVC) before redeploying, or the running SQL Server will still" >&2
    echo "have the old password baked in and login will fail." >&2
  fi
fi

CONN="Server=sqlserver,1433;Database=UrlShortener;User Id=sa;Password=${PASSWORD};TrustServerCertificate=True"

kubectl -n url-shortener create secret generic url-shortener-secrets \
  --from-literal=MSSQL_SA_PASSWORD="${PASSWORD}" \
  --from-literal=ConnectionStrings__Default="${CONN}" \
  --dry-run=client -o yaml >"${DIR}/secrets.yaml"

echo "Wrote ${DIR}/secrets.yaml"
