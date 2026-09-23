// Package database 负责 SQLite 数据库的初始化、建表与种子数据。
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init 打开（必要时创建）SQLite 数据库文件，建表并填充种子数据。
func Init(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // SQLite 单写者，避免并发写锁冲突
	DB = db

	if err := migrate(); err != nil {
		return err
	}
	if err := seed(); err != nil {
		return err
	}
	log.Printf("数据库就绪: %s", dbPath)
	return nil
}

// Close 关闭数据库连接。
func Close() {
	if DB != nil {
		_ = DB.Close()
	}
}

// migrate 创建数据表。
func migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	username   TEXT    NOT NULL UNIQUE,
	password   TEXT    NOT NULL,
	role       TEXT    NOT NULL DEFAULT 'user',
	created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS devices (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	device_no    TEXT    NOT NULL UNIQUE,
	name         TEXT    NOT NULL,
	type         TEXT    NOT NULL,
	location     TEXT    NOT NULL,
	status       TEXT    NOT NULL DEFAULT '在线',
	firmware     TEXT    NOT NULL DEFAULT 'v1.0.0',
	last_seen_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS daily_metrics (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	day      TEXT    NOT NULL UNIQUE,
	messages INTEGER NOT NULL DEFAULT 0,
	devices  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS alerts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	alert_no   TEXT    NOT NULL UNIQUE,
	device_no  TEXT    NOT NULL,
	metric     TEXT    NOT NULL,
	level      TEXT    NOT NULL DEFAULT '提示',
	value      TEXT    NOT NULL DEFAULT '',
	status     TEXT    NOT NULL DEFAULT '未处理',
	created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS tokens (
	token      TEXT    PRIMARY KEY,
	user_id    INTEGER NOT NULL,
	expires_at TEXT    NOT NULL
);
`
	if _, err := DB.Exec(schema); err != nil {
		return err
	}
	dropLegacyUserColumns()
	return nil
}

// dropLegacyUserColumns 清理历史库中 users 表已废弃的 nickname / email / status 列。
func dropLegacyUserColumns() {
	rows, err := DB.Query("PRAGMA table_info(users)")
	if err != nil {
		return
	}
	defer rows.Close()

	existing := map[string]bool{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, typ string
		var dflt any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return
		}
		existing[name] = true
	}

	for _, col := range []string{"nickname", "email", "status"} {
		if !existing[col] {
			continue
		}
		if _, err := DB.Exec("ALTER TABLE users DROP COLUMN " + col); err != nil {
			log.Printf("清理历史列 users.%s 失败: %v", col, err)
		}
	}
}

// seed 在表为空时写入演示数据。
func seed() error {
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := seedUsers(); err != nil {
			return err
		}
	}
	if err := DB.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := seedDevices(); err != nil {
			return err
		}
	}
	if err := DB.QueryRow("SELECT COUNT(*) FROM daily_metrics").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := seedMetrics(); err != nil {
			return err
		}
	}
	if err := DB.QueryRow("SELECT COUNT(*) FROM alerts").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := seedAlerts(); err != nil {
			return err
		}
	}
	return nil
}
