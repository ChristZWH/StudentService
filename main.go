package main

import (
	"StudentService/database"
	"StudentService/handleers"
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
	if err := database.InitDatabases(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
		return
	}
	defer database.CloseDatabases() // 确保关闭所有连接

	// 路由
	router := gin.Default()

	// 添加日志中间件
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("这个%s %s %d %s", c.Request.Method, c.Request.URL, c.Writer.Status(), duration)
	})

	router.POST("/login", handleers.LoginHandler)
	// 路由配置restful风
	studentRoutes := router.Group("/students")
	studentRoutes.Use(handleers.JWTAuthMiddleware()) // 应用JWT中间件
	{
		studentRoutes.GET("/", handleers.ListStudents)
		studentRoutes.POST("/", handleers.CreateStudent)
		studentRoutes.GET("/:id", handleers.GetStudent)
		studentRoutes.PUT("/:id", handleers.UpdateStudent)
		studentRoutes.DELETE("/:id", handleers.DeleteStudent)
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	log.Println("服务已关闭")
}
