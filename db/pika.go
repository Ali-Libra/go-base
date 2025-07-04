package db

import (
	"context"
	"go-base/logger"
	"sync"

	"github.com/redis/go-redis/v9"
)

type PikaMgr struct {
	client *redis.Client
}

var (
	instance *PikaMgr
	once     sync.Once
)

func DefaultPikaMgr() *PikaMgr {
	once.Do(func() {
		instance = &PikaMgr{}
	})
	return instance
}

func NewPikaMgr() *PikaMgr {
	return &PikaMgr{}
}

func (mgr *PikaMgr) Init(addr string) bool {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis 地址
		Password: "",   // 无密码则留空
		DB:       0,    // 使用默认 DB
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("❌ Redis 连接失败:", err)
		return false
	}

	mgr.client = rdb
	return true
}

func (p *PikaMgr) Close() {
	p.client.Close()
}

func (p *PikaMgr) GetClient() *redis.Client {
	return p.client
}
