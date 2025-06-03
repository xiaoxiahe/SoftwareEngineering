package database

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// NewConnection 创建新的数据库连接
func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	db.SetMaxOpenConns(25)                 // 最大连接数
	db.SetMaxIdleConns(10)                 // 最大空闲连接数
	db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期
	db.SetConnMaxIdleTime(3 * time.Minute) // 空闲连接最大生命周期

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// RunMigrations 运行数据库迁移
func RunMigrations(cfg config.DatabaseConfig) error {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("无法连接数据库进行迁移: %v", err)
	}
	defer db.Close()

	// 确保数据库连接有效
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("创建迁移驱动实例失败: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %v", err)
	}

	// 获取当前数据库版本
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取数据库版本失败: %v", err)
	}

	// 如果数据库处于dirty状态，尝试强制设置为特定版本然后重新执行迁移
	if dirty {
		fmt.Printf("数据库处于dirty状态，版本: %d，尝试修复...\n", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("修复数据库dirty状态失败: %v", err)
		}
		fmt.Printf("数据库状态已修复，将从版本 %d 继续迁移\n", version)
	} else if err == migrate.ErrNilVersion {
		fmt.Println("首次执行数据库迁移")
	} else {
		fmt.Printf("当前数据库版本: %d\n", version)
	}

	// 执行迁移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行数据库迁移失败: %v", err)
	} else if err == migrate.ErrNoChange {
		fmt.Println("数据库已是最新版本，无需迁移")
	} else {
		newVersion, _, _ := m.Version()
		fmt.Printf("数据库迁移成功，当前版本: %d\n", newVersion)
	}

	return nil
}
