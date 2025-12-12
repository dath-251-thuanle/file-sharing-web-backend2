# Chạy server development
server:
	go run ./cmd/server/main.go

# Reset database (xóa data + khởi động lại)
docker-reset:
	docker compose down -v
	docker compose up -d

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
