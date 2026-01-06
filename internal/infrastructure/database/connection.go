package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db error: %w", err)
	}

	DB.SetMaxIdleConns(3)                   // Số kết nối nhàn rỗi tối đa
	DB.SetMaxOpenConns(30)                  // Số kết nối tôi đa
	DB.SetConnMaxLifetime(30 * time.Minute) // Đóng kết nối sau 30 phút
	DB.SetConnMaxIdleTime(5 * time.Minute)  // Đóng kết nối nhàn rỗi sau 5 phút

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		DB.Close()
		return nil, fmt.Errorf("DB ping error: %w", err)
	}

	log.Println("Connected")

	return DB, nil
}
