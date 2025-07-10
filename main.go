package main

import (
	"StudentService/database"
	handlers "StudentService/handleers"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Gin version:", gin.Version)
	// 初始化数据库
	if err := database.InitDatabases(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDatabases() // 确保关闭所有连接

	// 路由
	router := gin.Default()

	// 路由配置restful风
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
	go func() {
		log.Println("服务启动，监听端口 8080...")
		// func (srv *http.Server) ListenAndServe() error
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器错误: %v", err)
		}
	}()

	// 关闭处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("接收到关闭信号，开始关闭...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	log.Println("服务已关闭")
	//最后还会调用栈main的上一个函数：database.CloseDatabases()
}
