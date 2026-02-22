package db

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Ali-Libra/go-base/env"
	"github.com/Ali-Libra/go-base/logger"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/ssh"
)

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

var (
	instance *RedisMgr
	once     sync.Once
)

func DefaultPikaMgr() *RedisMgr {
	once.Do(func() {
		instance = &RedisMgr{}
	})
	return instance
}

func NewRedisMgr() *RedisMgr {
	return &RedisMgr{}
}

type RedisMgr struct {
	client *redis.Client
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

func CreateRedisClient(addr string, useSSH bool) *redis.Client {
	redis_addr := ""
	if addr != "" {
		redis_addr = addr
	} else {
		redis_addr = env.GetEnv("REDIS_ADDR")
	}

	var dialer func(ctx context.Context, network, addr string) (net.Conn, error)
	if useSSH {
		sshClient, err := buildSSHClient()
		if err != nil {
			logger.Error("❌ SSH 隧道连接失败: %v", err)
			return nil
		}
		dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := sshClient.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return &noDeadlineConn{Conn: conn}, nil
		}
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_addr, // Redis 地址
		Password: "",         // 无密码则留空
		DB:       0,          // 使用默认 DB
		Dialer:   dialer,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("❌ Redis 连接失败: %v", err)
		return nil
	}

	return rdb
}

type noDeadlineConn struct {
	net.Conn
}

func (c *noDeadlineConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *noDeadlineConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *noDeadlineConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func buildSSHClient() (*ssh.Client, error) {
	sshHost := env.GetEnv("SSH_HOST")
	if sshHost == "" {
		sshHost = env.GetEnv("REMOTE_ADDR")
	}
	if sshHost == "" {
		return nil, fmt.Errorf("SSH_HOST is empty")
	}

	sshPort := env.GetEnv("SSH_PORT")
	if sshPort == "" {
		sshPort = "22"
	}
	sshUser := env.GetEnv("SSH_USER")
	if sshUser == "" {
		sshUser = "ubuntu"
	}

	keyPath := env.GetEnv("SSH_KEY_PATH")
	if keyPath == "" {
		keyName := env.GetEnv("SSH_KEY_NAME")
		if keyName == "" {
			keyName = "id_rsa"
		}
		homeDir := os.Getenv("USERPROFILE")
		if homeDir == "" {
			homeDir = os.Getenv("HOME")
		}
		keyPath = filepath.Join(homeDir, ".ssh", keyName)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取SSH密钥失败: %w %s", err, keyPath)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("解析SSH密钥失败: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%s", sshHost, sshPort)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	return client, nil
}
