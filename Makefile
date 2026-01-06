LOCAL = docker-compose.yml -f docker-compose.local.yml

# Khởi động docker cho LOCAL
docker-up:
	docker compose -f $(LOCAL) up -d --build

# Stop containers
docker-down:
	docker compose -f $(LOCAL) down

# Reset DB LOCAL
docker-reset:
	docker compose -f $(LOCAL) down -v
	docker compose -f $(LOCAL) up -d --build

# Xem logs API
docker-logs:
	docker compose logs -f api

# Chạy tests
test:
	go test ./...

# Xóa build artifacts
clean:
	rm -rf bin/

# Tải dependencies
deps:
	go mod download
	go mod tidy

.PHONY: server docker-reset docker-logs test clean deps
