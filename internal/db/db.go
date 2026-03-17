package db

import (
	"database/sql"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// PoolConfig 连接池配置。字段为 0 时从 env 读取，env 未设则用默认值
type PoolConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大存活时间
}

// OpenMySQL 从环境变量构建 DSN 并建立 MySQL 连接池
//
// DSN：BASE_MYSQL_URL, BASE_MYSQL_PORT, BASE_MYSQL_USER, BASE_MYSQL_PASS, BASE_MYSQL_DATABASE
// 连接池（可选）：MYSQL_MAX_OPEN_CONNS, MYSQL_MAX_IDLE_CONNS, MYSQL_CONN_MAX_LIFETIME（如 1h）
func OpenMySQL(cfg PoolConfig) (*sql.DB, error) {
	rawURL := os.Getenv("BASE_MYSQL_URL")
	if rawURL == "" {
		rawURL = "localhost"
	}
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		rawURL = u.Host
	} else {
		rawURL = strings.TrimPrefix(rawURL, "http://")
		rawURL = strings.TrimPrefix(rawURL, "https://")
	}
	port := os.Getenv("BASE_MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("BASE_MYSQL_USER")
	pass := os.Getenv("BASE_MYSQL_PASS")
	dbname := os.Getenv("BASE_MYSQL_DATABASE")
	if dbname == "" {
		dbname = "media_operator"
	}
	dsn := user + ":" + pass + "@tcp(" + rawURL + ":" + port + ")/" + dbname + "?charset=utf8mb4&parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 连接池：优先用 cfg，否则从 env 读，最后默认
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		if n := os.Getenv("MYSQL_MAX_OPEN_CONNS"); n != "" {
			maxOpen, _ = strconv.Atoi(n)
		}
		if maxOpen <= 0 {
			maxOpen = 25
		}
	}
	db.SetMaxOpenConns(maxOpen)
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		if n := os.Getenv("MYSQL_MAX_IDLE_CONNS"); n != "" {
			maxIdle, _ = strconv.Atoi(n)
		}
		if maxIdle <= 0 {
			maxIdle = 10
		}
	}
	db.SetMaxIdleConns(maxIdle)
	connLife := cfg.ConnMaxLifetime
	if connLife <= 0 {
		if s := os.Getenv("MYSQL_CONN_MAX_LIFETIME"); s != "" {
			connLife, _ = time.ParseDuration(s)
		}
		if connLife <= 0 {
			connLife = 5 * time.Minute
		}
	}
	db.SetConnMaxLifetime(connLife)
	return db, nil
}
