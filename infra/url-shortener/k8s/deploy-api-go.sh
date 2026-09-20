#!/usr/bin/env bash
# Builds the Go API image and (re)deploys it with a given CPU/memory
# ceiling, as a separate Service (api-go) alongside the .NET one (api), so
# both can be load-tested under identical limits and compared:
#   API_CPUS=0.5 API_MEMORY=256Mi ./deploy-api-go.sh
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DIR}/../../.." && pwd)"

export API_REPLICAS="${API_REPLICAS:-1}"
export API_CPUS="${API_CPUS:-1}"
export API_CPUS_REQUEST="${API_CPUS_REQUEST:-$(awk "BEGIN{printf \"%.3f\", ${API_CPUS}/2}")}"
export API_MEMORY="${API_MEMORY:-512Mi}"
export API_MEMORY_REQUEST="${API_MEMORY_REQUEST:-${API_MEMORY}}"
export API_IMAGE="${API_IMAGE:-url-shortener-api-go:local}"

echo "Building ${API_IMAGE} ..."
docker build -t "${API_IMAGE}" "${REPO_ROOT}/backend/GoApi"

echo "Deploying api-go with cpu<=${API_CPUS}, memory<=${API_MEMORY} (replicas=${API_REPLICAS}) ..."
envsubst <"${DIR}/api-go.yaml.tmpl" | kubectl apply -f -

kubectl -n url-shortener rollout restart deployment/api-go
kubectl -n url-shortener rollout status deployment/api-go
