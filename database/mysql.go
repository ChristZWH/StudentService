package database

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// var DB *sql.DB
var GormDB *gorm.DB

func InitMySQL() error {
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	// 测试阶段直接设置密码
	pass := "ZWH20050512"
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/student_db?parseTime=true&charset=utf8mb4&timeout=5s", user, pass)

	var err error
	// DB, err = sql.Open("mysql", dsn)
	// if err != nil {
	// 	return fmt.Errorf("数据库连接失败: %w", err)
	// }
	// 使用 GORM 打开连接
	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("GORM 数据库连接失败: %w", err)
	}
	// 配置 GORM 底层的 database/sql 连接池
	DB, err := GormDB.DB() // 获取底层的 *sql.DB 对象
	if err != nil {
		return fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}
	// 配置连接池
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(10 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// log.Println("MySQL连接成功")
	log.Println("MySQL(GORM)连接成功")
	return nil
}

func Close() {
	// if DB != nil {
	// 	if err := DB.Close(); err != nil {
	// 		log.Printf("关闭MySQL连接失败: %v", err)
	// 	} else {
	// 		log.Println("MySQL连接已关闭")
	// 	}
	// }
	if GormDB != nil {
		sqlDB, err := GormDB.DB() // 获取底层 *sql.DB
		if err != nil {
			log.Printf("获取底层 sql.DB 以关闭失败: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil { // 关闭底层 *sql.DB
			log.Printf("关闭 MySQL 连接失败: %v", err)
		} else {
			log.Println("MySQL 连接已关闭")
		}
	}
}

func CreateGorm() (*gorm.DB, error) {
	user := "root"
	pass := "ZWH20050512"
	dns := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/student_db?parseTime=true&charset=utf8mb4&timeout=5s", user, pass)
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	return db, err
}
