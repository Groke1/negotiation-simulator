.PHONY: start stop logs sqlc test

start:
	docker compose -f docker-compose.yml up -d --build

stop:
	docker compose -f docker-compose.yml down -v

logs:
	docker compose logs -f arena

sqlc:
	cd backend && sqlc generate

test:
	cd backend && go test ./...
