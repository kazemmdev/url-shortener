# Kubernetes: resource-limited API + k6

The same experiment as the root `makefile`'s `api-docker` / `loadtest` / `stats`
targets, but on a real cluster instead of `docker --cpus`. Kubernetes enforces
CPU limits with a CFS quota (the process gets throttled in fixed time slices
once it exceeds its share), which behaves more like a genuinely small VM than
Docker's limiter does — useful for seeing how much headroom the Redis cache
actually buys the API when the box is small.

Uses whatever cluster your current `kubectl` context points at. On Docker
Desktop, enable Kubernetes in its settings and the `docker-desktop` context
works out of the box — no image registry needed, since the cluster shares
the local Docker image store.

## One-time setup

```bash
make k8s-up
```

Creates the `url-shortener` namespace, generates `secrets.yaml` (gitignored —
random `MSSQL_SA_PASSWORD`), and deploys SQL Server + Redis. Wait for the
rollout to finish (the target does this for you) before deploying the API.

## Deploy the API with a resource ceiling

```bash
make k8s-api API_CPUS=1 API_MEMORY=512Mi     # "normal" server
make k8s-api API_CPUS=0.25 API_MEMORY=256Mi  # constrained server
```

Builds `backend/DotnetApi` into `url-shortener-api:local` and redeploys.
Re-run with different values any time — it always forces a rollout so the
freshly built image is picked up even though the tag doesn't change.

## Load test

```bash
make k8s-loadtest PROFILE=stress TARGET_RPS=3000
```

Runs `loadtest/url-shortener.js` as a k6 Job inside the cluster, hitting the
`api` Service directly (no port-forward in the loop). Streams logs to your
terminal; k6's summary (RPS achieved, p95/p99 latency, error rate) prints at
the end. Past runs stick around as completed Jobs for an hour
(`kubectl -n url-shortener get jobs`) if you want to `kubectl logs` them again.

## Watch resource usage live

`kubectl top` needs metrics-server, which Docker Desktop doesn't ship by
default:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl -n kube-system patch deployment metrics-server --type=json \
  -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
```

Then, while a load test runs:

```bash
make k8s-stats
```

## Suggested experiment

1. `make k8s-api API_CPUS=0.25 API_MEMORY=256Mi`
2. `make k8s-loadtest PROFILE=stress TARGET_RPS=1000` and note p95/p99 for
   `redirect` (cache hits after the first read) vs `create` (always hits SQL
   Server), plus how quickly the API pod gets CPU-throttled
   (`kubectl -n url-shortener describe pod -l app=api` shows throttling
   restarts if any; `k8s-stats` shows sustained CPU% against the limit).
3. Compare against the same profile run with a more generous ceiling
   (`API_CPUS=1 API_MEMORY=512Mi`) to see how much of that gap Redis is
   already covering, and how much is SQL Server contention.

## Reset

```bash
make k8s-down
```

Deletes the whole `url-shortener` namespace (including the PVC — SQL Server
data is not preserved).
