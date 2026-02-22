package db

import (
	"context"
	"fmt"
	"net"

	"github.com/Ali-Libra/go-base/env"
	"github.com/Ali-Libra/go-base/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePostgressClient(addr string, useSSH bool) *pgxpool.Pool {
	postgre_addr := ""
	if addr != "" {
		postgre_addr = addr
	} else {
		postgre_addr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			env.GetEnv("DB_HOST"),
			env.GetEnv("DB_PORT"),
			env.GetEnv("DB_USER"),
			env.GetEnv("DB_PASSWORD"),
			env.GetEnv("DB_NAME"),
		)
	}
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(postgre_addr)
	if err != nil {
		logger.Error("PostgreSQL配置解析失败: %v", err)
		return nil
	}
	if useSSH {
		sshClient, err := buildSSHClient()
		if err != nil {
			logger.Error("❌ SSH 隧道连接失败: %v", err)
			return nil
		}
		config.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := sshClient.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return &noDeadlineConn{Conn: conn}, nil
		}
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
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
