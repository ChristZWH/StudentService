/*
   通过一下步骤查询数据库名称：
   sudo mysql -u root -p                 登录数据库
   SELECT User, Host FROM mysql.user;    查询所有用户
   SELECT CURRENT_USER();                查看当前用户

   查看绑定地址：SHOW VARIABLES LIKE 'bind_address';
   查看MySQl端口：SHOW VARIABLES LIKE 'port';
*/
// user := "root"

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitMySQL() error {
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root" // 默认用户
	}
	// 测试阶段直接设置密码
	pass := "ZWH20050512"
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/student_db?parseTime=true&charset=utf8mb4&timeout=5s", user, pass)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	// 配置连接池
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(20 * time.Minute)
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("MySQL连接成功")
	return nil
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("关闭MySQL连接失败: %v", err)
		} else {
			log.Println("MySQL连接已关闭")
		}
	}
}
