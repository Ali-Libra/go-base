package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ali-Libra/go-base/env"
	"github.com/Ali-Libra/go-base/logger"
	_ "modernc.org/sqlite"
)

func CreateSqliteClient(dbPath string) *sql.DB {
	sqlitePath := dbPath
	if sqlitePath == "" {
		sqlitePath = env.GetEnv("SQLITE_PATH")
	}
	if sqlitePath == "" {
		sqlitePath = "./data/app.db"
	}

	dir := filepath.Dir(sqlitePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			logger.Error("创建SQLite目录失败: %v", err)
			return nil
		}
	}

	dsn := fmt.Sprintf("file:%s", sqlitePath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		logger.Error("SQLite打开失败: %v", err)
		return nil
	}

	if err = db.Ping(); err != nil {
		logger.Error("SQLite连接失败: %v", err)
		_ = db.Close()
		return nil
	}

	return db
}
