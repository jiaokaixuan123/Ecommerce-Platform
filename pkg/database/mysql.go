package database

import (
	"time"

	"github.com/ecommerce-platform/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(cfg *config.MySQLConfig) (*gorm.DB, error) {
	// 1. 基于 DSN 初始化 GORM DB 实例
	// mysql.Open(cfg.DSN)：通过 DSN 连接字符串创建 MySQL 驱动
	// &gorm.Config{}：GORM 配置，这里设置日志级别为 Info（打印 SQL 等日志）
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开启 Info 级日志（包含 SQL 执行日志）
	})
	if err != nil {
		return nil, err // 驱动初始化失败，返回错误
	}

	// 2. 获取底层的 sql.DB 实例（GORM 封装了标准库的 sql.DB）
	// 用于配置连接池、超时等底层参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err // 获取 sql.DB 失败，返回错误
	}

	// 3. 配置数据库连接池（关键优化项）
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn) // 设置最大打开连接数（同时使用的连接数上限）
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn) // 设置最大空闲连接数（空闲时保持的连接数）
	sqlDB.SetConnMaxLifetime(time.Hour)    // 设置连接最大存活时间（避免长期占用连接）

	// 4. 返回初始化完成的 GORM DB 实例
	return db, nil
}
