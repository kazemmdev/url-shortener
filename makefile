.PHONY: up infra api-dotnet api-docker loadtest stats k8s-up k8s-api k8s-api-go k8s-loadtest k8s-stats k8s-down

up: infra api-docker

infra:
	docker compose up -d sqlserver redis

api-dotnet:
	cd backend/DotnetApi && dotnet run

# Run the API in docker with limited resources: make api-docker API_CPUS=0.5 API_MEMORY=256M
api-docker:
	docker compose up -d --build api

# Load test with k6: make loadtest PROFILE=stress K6_ARGS="-e TARGET_RPS=5000"
PROFILE ?= stress
K6_ARGS ?= -e TARGET_RPS=50000

loadtest:
	docker compose up -d --force-recreate api && MSYS_NO_PATHCONV=1 docker compose --profile loadtest run --rm k6 run -e PROFILE=$(PROFILE) $(K6_ARGS) --summary-export=/results/summary-$(PROFILE).json /scripts/url-shortener.js

# Watch CPU / memory of the API and database while a test runs
stats:
	docker stats url-shortener-api sqlserver

# --- Kubernetes (infra/url-shortener/k8s): same experiment, real CFS-quota throttling ---
K8S_DIR := infra/url-shortener/k8s

# First-time (or full reset): namespace, secrets, sqlserver, redis
k8s-up:
	$(K8S_DIR)/create-secrets.sh
	kubectl apply -k $(K8S_DIR)
	kubectl -n url-shortener rollout status deployment/sqlserver

# Build + (re)deploy the .NET API with a resource ceiling: make k8s-api API_CPUS=0.5 API_MEMORY=512Mi
k8s-api:
	$(K8S_DIR)/deploy-api.sh

# Same, for the Go API (separate Service "api-go", same DB/Redis): make k8s-api-go API_CPUS=0.5 API_MEMORY=256Mi
k8s-api-go:
	$(K8S_DIR)/deploy-api-go.sh

# Load test with k6 as an in-cluster Job: make k8s-loadtest PROFILE=stress TARGET_RPS=3000 TARGET_SERVICE=api-go
k8s-loadtest:
	$(K8S_DIR)/run-loadtest.sh

# Watch CPU / memory of the pods while a test runs (needs metrics-server, see k8s README)
k8s-stats:
	watch kubectl -n url-shortener top pod

k8s-down:
	kubectl delete namespace url-shortener --ignore-not-found
