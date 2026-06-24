package db

import (
	"cake/env"
	zlogger "cake/internal/pkg/logger"
	"context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// 全局DB单例
var dbInst *gorm.DB

func DbInst() *gorm.DB {
	return dbInst
}

// InitMysql  初始化MySQL连接
func InitMysql() {
	user := env.GetString("db.user")
	pass := env.GetString("db.pass")
	host := env.GetString("db.host")
	port := env.GetString("db.port")
	name := env.GetString("db.name")
	dsn := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), // 打印SQL
	})
	if err != nil {
		panic(err)
	}

	// 设置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	dbInst = db
	zlogger.Infof("Success	MySQL 连接成功")
}

func CloseMysql() {
	if dbInst == nil {
		zlogger.Error("MySQL未初始化，无需关闭")
		return
	}
	sqlDB, err := dbInst.DB()
	if err != nil {
		zlogger.Errorf("获取底层SQL DB失败: %v", err)
		return
	}

	// 设置全局操作超时，最多等待15秒
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 等待空闲连接释放，超时强制关闭
	done := make(chan struct{})
	go func() {
		_ = sqlDB.Close()
		close(done)
	}()

	select {
	case <-done:
		zlogger.Error("MySQL连接池正常关闭")
	case <-ctx.Done():
		zlogger.Error("MySQL关闭超时，强制终止数据库连接")
	}
}
