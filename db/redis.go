package db

import (
	"context"
	"go-base/logger"
	"sync"

	"github.com/redis/go-redis/v9"
)

type RedisMgr struct {
	client *redis.Client
}

var (
	redisInstance *RedisMgr
	redisOnce     sync.Once
)

func DefaultRedisMgr() *RedisMgr {
	redisOnce.Do(func() {
		redisInstance = &RedisMgr{}
	})
	return redisInstance
}

func NewRedisMgr() *RedisMgr {
	return &RedisMgr{}
}

func (mgr *RedisMgr) Init(addr string) bool {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis 地址
		Password: "",   // 无密码则留空
		DB:       0,    // 使用默认 DB
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("❌ Redis 连接失败: %v", err)
		return false
	}

	mgr.client = rdb
	return true
}

func (p *RedisMgr) Close() {
	p.client.Close()
}

func (p *RedisMgr) GetClient() *redis.Client {
	return p.client
}
