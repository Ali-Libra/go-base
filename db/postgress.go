package db

import (
	"context"
	"fmt"

	"github.com/Ali-Libra/go-base/env"
	"github.com/Ali-Libra/go-base/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePostgressClient() *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		env.GetEnv("DB_HOST"),
		env.GetEnv("DB_PORT"),
		env.GetEnv("DB_USER"),
		env.GetEnv("DB_PASSWORD"),
		env.GetEnv("DB_NAME"),
	)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("PostgreSQL连接失败: %v", err)
		return nil
	}
	// 测试连接
	err = pool.Ping(ctx)
	if err != nil {
		logger.Error("PostgreSQL ping失败: %v", err)
		return nil
	}
	return pool
}
