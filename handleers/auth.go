package handleers

import (
	"StudentService/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	authConfig = models.AuthConfig{
		JWTSecret:     "111",
		TokenDuration: 24 * time.Hour,
	}
)

// 登录处理
func LoginHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("登录请求解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 这里简化代码没有从数据库中查询
	if user.Username != "admin" || user.Password != "admin123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 创建JWTtoken
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authConfig.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "StudentService",
		},
	})

	tokenString, err := token.SignedString([]byte(authConfig.JWTSecret))
	if err != nil {
		log.Printf("生成JWT失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成signature失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// JWT验证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 移除"Bearer "前缀（如果存在）
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(authConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 将用户名存入上下文
		c.Set("username", claims.Username)
		c.Next()
	}
}
