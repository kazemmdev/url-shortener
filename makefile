.PHONY: up infra api-dotnet api-docker loadtest stats

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
