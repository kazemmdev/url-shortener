#!/usr/bin/env bash
# Builds the API image and (re)deploys it with a given CPU/memory ceiling,
# to simulate running on a smaller server:
#   ./deploy-api.sh                          # defaults below
#   API_CPUS=0.25 API_MEMORY=256Mi ./deploy-api.sh
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DIR}/../../.." && pwd)"

export API_REPLICAS="${API_REPLICAS:-1}"
export API_CPUS="${API_CPUS:-1}"
export API_CPUS_REQUEST="${API_CPUS_REQUEST:-$(awk "BEGIN{printf \"%.3f\", ${API_CPUS}/2}")}"
export API_MEMORY="${API_MEMORY:-512Mi}"
export API_MEMORY_REQUEST="${API_MEMORY_REQUEST:-${API_MEMORY}}"
export API_IMAGE="${API_IMAGE:-url-shortener-api:local}"

echo "Building ${API_IMAGE} ..."
docker build -t "${API_IMAGE}" "${REPO_ROOT}/backend/DotnetApi"

echo "Deploying api with cpu<=${API_CPUS}, memory<=${API_MEMORY} (replicas=${API_REPLICAS}) ..."
envsubst <"${DIR}/api.yaml.tmpl" | kubectl apply -f -

kubectl -n url-shortener rollout restart deployment/api
kubectl -n url-shortener rollout status deployment/api
