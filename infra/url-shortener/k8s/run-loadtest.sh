#!/usr/bin/env bash
# Runs the same k6 script as `make loadtest`, but as an in-cluster Job
# hitting an api Service directly:
#   ./run-loadtest.sh                                            # defaults below, targets api (.NET)
#   PROFILE=stress TARGET_RPS=3000 TARGET_SERVICE=api-go ./run-loadtest.sh
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${DIR}/../../.." && pwd)"

export PROFILE="${PROFILE:-stress}"
export TARGET_RPS="${TARGET_RPS:-3000}"
export WRITE_PERCENT="${WRITE_PERCENT:-10}"
export TARGET_SERVICE="${TARGET_SERVICE:-api}"
export RUN_ID="$(date +%Y%m%d-%H%M%S)"
JOB="k6-${TARGET_SERVICE}-${RUN_ID}"

kubectl -n url-shortener create configmap k6-script \
  --from-file="${REPO_ROOT}/loadtest/url-shortener.js" \
  --dry-run=client -o yaml | kubectl apply -f -

envsubst <"${DIR}/loadtest-job.yaml.tmpl" | kubectl apply -f -

echo "Waiting for job ${JOB} (target=${TARGET_SERVICE}) ..."
# Don't wait on condition=complete alone: k6 exits non-zero (and the Job
# lands in Failed, not Complete) whenever a threshold is crossed, which is a
# normal outcome here, not a script error. Wait for either terminal state.
kubectl -n url-shortener wait --for=jsonpath='{.status.conditions[?(@.status=="True")].type}'=Complete \
  "job/${JOB}" --timeout=15m >/dev/null 2>&1 &
COMPLETE_PID=$!
kubectl -n url-shortener wait --for=jsonpath='{.status.conditions[?(@.status=="True")].type}'=Failed \
  "job/${JOB}" --timeout=15m >/dev/null 2>&1 &
FAILED_PID=$!

kubectl -n url-shortener logs -f "job/${JOB}" || true

# Whichever terminal state already happened returns immediately; kill the other.
wait -n "${COMPLETE_PID}" "${FAILED_PID}" 2>/dev/null || true
kill "${COMPLETE_PID}" "${FAILED_PID}" 2>/dev/null || true

SUCCEEDED="$(kubectl -n url-shortener get "job/${JOB}" -o jsonpath='{.status.succeeded}')"
if [[ "${SUCCEEDED}" == "1" ]]; then
  echo "${JOB}: all thresholds passed"
else
  echo "${JOB}: finished with threshold failures (see output above) — this is a valid, informative result, not a script error"
fi
