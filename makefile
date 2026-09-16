
infra:
	docker compose up -d sqlserver redis

api-dotnet:
	cd backend/DotnetApi && dotnet run