import-db:
	docker exec -i postgres-db psql -U haixon -d file-sharing < ./internal/infrastructure/database/init.sql
export-db:
	docker exec -i postgres-db pg_dump -U haixon -d file-sharing > ./internal/infrastructure/database/backup.sql
server:
	go run ./cmd/server/main.go
build:
	go build -o bin/myapp.exe ./cmd/server
run-binary:
	./bin/myapp.exe
