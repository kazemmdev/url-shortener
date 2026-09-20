.PHONY: up infra api-dotnet api-docker

up: infra api-docker

infra:
	docker compose up -d sqlserver redis

api-dotnet:
	cd backend/DotnetApi && dotnet run

# Run the API in docker with limited resources: make api-docker API_CPUS=0.5 API_MEMORY=256M
api-docker:
	docker compose up -d --build api
