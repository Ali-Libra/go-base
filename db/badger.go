package db

import (
	"errors"
	"os"

	"github.com/Ali-Libra/go-base/env"
	"github.com/Ali-Libra/go-base/logger"
	"github.com/dgraph-io/badger/v4"
)

// ErrNotFound is returned by BadgerClient.Get when the key does not exist.
var ErrNotFound = errors.New("key not found")

type BadgerClient struct {
	db  *badger.DB
	seq *badger.Sequence
}

func CreateBadgerClient(path string) *BadgerClient {
	if path == "" {
		path = env.GetEnv("BADGER_PATH")
	}
	if path == "" {
		path = "./data/badger"
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		logger.Error("创建BadgerDB目录失败: %v", err)
		return nil
	}

	opts := badger.DefaultOptions(path).WithLogger(nil)
	bdb, err := badger.Open(opts)
	if err != nil {
		logger.Error("BadgerDB打开失败: %v", err)
		return nil
	}

	seq, err := bdb.GetSequence([]byte("__id_seq__"), 100)
	if err != nil {
		logger.Error("BadgerDB序列初始化失败: %v", err)
		_ = bdb.Close()
		return nil
	}

	return &BadgerClient{db: bdb, seq: seq}
}

// NextID returns a monotonically increasing unique uint64 ID.
func (c *BadgerClient) NextID() (uint64, error) {
	return c.seq.Next()
}

// Get returns the value for key. Returns ErrNotFound if key does not exist.
func (c *BadgerClient) Get(key []byte) ([]byte, error) {
	var val []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

// Set stores the key-value pair.
func (c *BadgerClient) Set(key, value []byte) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Delete removes a key.
func (c *BadgerClient) Delete(key []byte) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (c *BadgerClient) Close() {
	if c.seq != nil {
		_ = c.seq.Release()
	}
	if c.db != nil {
		_ = c.db.Close()
	}
}
