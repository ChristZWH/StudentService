package main

import (
	"StudentService/database"
	handlers "StudentService/handleers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	if err := database.InitMySQL(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close() // 确保关闭连接

	// 创建Gin路由
	router := gin.Default()

	// 设置可信代理（生产环境必需）
	router.SetTrustedProxies([]string{"127.0.0.1"})

	// RESTful路由配置
	studentRoutes := router.Group("/students")
	{
		studentRoutes.GET("/", handlers.ListStudents)
		studentRoutes.POST("/", handlers.CreateStudent)
		studentRoutes.GET("/:id", handlers.GetStudent)
		studentRoutes.PUT("/:id", handlers.UpdateStudent)
		studentRoutes.DELETE("/:id", handlers.DeleteStudent)
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Println("服务启动，监听端口 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器错误: %v", err)
		}
	}()

	// 优雅关闭处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("接收到关闭信号，开始关闭...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 停止监控
	database.StopMonitor()

	// 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	log.Println("服务已关闭")
}
