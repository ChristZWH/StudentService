package database

import "log"

// InitDatabases 初始化所有数据库
func InitDatabases() error {
	log.Println("初始化数据库中：")

	if err := InitMySQL(); err != nil {
		log.Printf("MySQL初始化失败: %v", err)
		return err
	}
	log.Println("MySQL初始化成功")

	if err := InitRedis(); err != nil {
		log.Printf("Redis初始化失败: %v", err)
		Close()
		return err
	}
	log.Println("Redis初始化成功")

	log.Println("所有数据库初始化完成")
	return nil
}

// CloseDatabases 关闭所有数据库连接
func CloseDatabases() {
	log.Println("关闭数据库")
	Close()
	CloseRedis()
	log.Println("所有数据库连接已关闭")
}
