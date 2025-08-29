package main

import (
	"StudentService/database"
	"StudentService/handleers"
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	if err := database.InitDatabases(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
		return
	}
	defer database.CloseDatabases()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	router := gin.Default()
	// 添加CORS中间件 (允许所有来源的跨域请求)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("请求：%s %s ; 状态：%d ; 持续时间： %s ; 客户端：%s", c.Request.Method, c.Request.URL, c.Writer.Status(), duration, c.ClientIP())
	})
	router.RedirectTrailingSlash = false // 禁用末尾斜杠重定向
	router.RedirectFixedPath = false     // 禁用路径规范化
	router.HandleMethodNotAllowed = true // 确保方法不匹配返回 405 而非重定向
	// router.POST("/login", handleers.LoginHandler)
	studentRoutes := router.Group("/students")
	// studentRoutes.Use(handleers.JWTAuthMiddleware()) // 应用JWT中间件
	{
		studentRoutes.GET("", handleers.ListStudents)
		studentRoutes.POST("", handleers.CreateStudent)
		studentRoutes.GET("/:id", handleers.GetStudent)
		studentRoutes.PUT("/:id", handleers.UpdateStudent)
		studentRoutes.DELETE("/:id", handleers.DeleteStudent)
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	srv.SetKeepAlivesEnabled(true)
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
