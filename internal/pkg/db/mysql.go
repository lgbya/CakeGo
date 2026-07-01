package db

import (
	"cake/env"
	zlogger "cake/internal/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// execAllSqlFileRaw 用原生sql.DB顺序执行所有sql文件
var sqlCommentRegex = regexp.MustCompile(`--[^\n]*`)

// 全局DB单例
var dbInst *gorm.DB

func DbInst() *gorm.DB {
	return dbInst
}

// InitMysql 初始化MySQL
// 逻辑：判断库是否存在，不存在则执行 ./sql 下所有SQL脚本（脚本里包含建库+建表语句）
func InitMysql() {
	user := env.GetString("db.user")
	pass := env.GetString("db.pass")
	host := env.GetString("db.host")
	port := env.GetString("db.port")
	name := env.GetString("db.name")

	// 1. 先连接系统默认mysql库，用来判断业务库是否存在
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port)
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		log.Fatal("连接mysql系统库失败: %w", err)
	}
	defer rootDB.Close()

	if err = rootDB.Ping(); err != nil {
		log.Fatal("mysql系统库连通失败: %w", err)
	}

	// 2. 判断业务库是否存在
	dbExists, err := databaseExists(rootDB, name)
	if err != nil {
		log.Fatal("无法判断数据库是否存在 ", err)
	}

	// 3. 库不存在：执行sql目录下全部初始化脚本（脚本内含建库语句）
	if !dbExists {
		zlogger.Infof("数据库[%s]不存在，开始执行sql目录初始化脚本", name)
		// 执行脚本时依然用系统库连接执行，脚本里会CREATE DATABASE
		if err = execAllSqlFileRaw(rootDB, "../../sql", name); err != nil {
			panic(fmt.Errorf("执行初始化SQL脚本失败: %w", err))
		}
		zlogger.Info("所有初始化SQL脚本执行完成")
	} else {
		zlogger.Infof("数据库[%s]已存在，跳过初始化SQL执行", name)
	}

	// 4. 正常连接业务库初始化GORM
	businessDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)
	db, err := gorm.Open(mysql.Open(businessDSN), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(fmt.Errorf("GORM连接业务库失败: %w", err))
	}

	// 设置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	dbInst = db
	zlogger.Infof("Success	MySQL 连接成功")
}

// databaseExists 判断数据库是否存在
func databaseExists(db *sql.DB, dbName string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?`
	err := db.QueryRow(query, dbName).Scan(&count)
	return count > 0, err
}

// targetDBName 从配置读取：game_db_1_1
func execAllSqlFileRaw(db *sql.DB, sqlDir string, targetDBName string) error {
	files, err := filepath.Glob(filepath.Join(sqlDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("遍历sql目录失败: %w", err)
	}
	if len(files) == 0 {
		zlogger.Warn("sql目录下未找到任何.sql初始化脚本")
		return nil
	}

	sort.Strings(files)

	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件 %s 失败: %w", filePath, err)
		}

		sqlText := string(data)
		// 核心：全局替换写死的game_db为 game_db_1_1
		sqlText = strings.ReplaceAll(sqlText, "game_db", targetDBName)

		// 移除单行注释
		sqlText = sqlCommentRegex.ReplaceAllString(sqlText, "")
		stmtList := strings.Split(sqlText, ";")

		for _, stmt := range stmtList {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("文件:%s 执行失败, SQL:\n%s\n错误:%w", filePath, stmt, err)
			}
		}
		zlogger.Infof("执行成功: %s", filepath.Base(filePath))
	}
	return nil
}

// GetDB 获取全局GORM实例
func GetDB() *gorm.DB {
	return dbInst
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
