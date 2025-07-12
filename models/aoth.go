package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 用户模型
type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// JWT声明
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 配置结构
type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
}
