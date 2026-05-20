package middleware

import (
	"net/http"
	"strings"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			model.Fail(c, http.StatusUnauthorized, "未认证")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			model.Fail(c, http.StatusUnauthorized, "令牌无效或已过期")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			model.Fail(c, http.StatusUnauthorized, "令牌解析失败")
			return
		}

		userID := uint(claims["user_id"].(float64))
		isAdmin, _ := claims["is_adminer"].(bool)

		// Verify user still exists and is active
		var user model.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "用户不存在")
			return
		}
		if user.Status == 0 {
			model.Fail(c, http.StatusForbidden, "账号已被冻结")
			return
		}

		c.Set("user_id", userID)
		c.Set("is_adminer", isAdmin)
		c.Next()
	}
}

// OptionalAuth 可选认证：有 token 则提取用户信息，无 token 则放行
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		userID := uint(claims["user_id"].(float64))
		isAdmin, _ := claims["is_adminer"].(bool)

		// Verify user still exists and is active
		var user model.User
		if err := config.DB.First(&user, userID).Error; err != nil || user.Status == 0 {
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Set("is_adminer", isAdmin)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_adminer")
		if !exists || !isAdmin.(bool) {
			model.Fail(c, http.StatusForbidden, "无管理员权限")
			return
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Api-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ApiKeyAuth 支持通过 Header X-Api-Key 或 Query 参数 ?api_key=xxx 认证
func ApiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Api-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		if apiKey == "" {
			model.Fail(c, http.StatusUnauthorized, "缺少 api_key 参数")
			c.Abort()
			return
		}

		var key model.ApiKey
		if err := config.DB.Where("key = ? AND revoked_at IS NULL", apiKey).First(&key).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "API Key 无效")
			c.Abort()
			return
		}

		// Verify user still exists and is active
		var user model.User
		if err := config.DB.First(&user, key.UserID).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "用户不存在")
			c.Abort()
			return
		}
		if user.Status == 0 {
			model.Fail(c, http.StatusForbidden, "账号已被冻结")
			c.Abort()
			return
		}

		// Update last_used
		now := time.Now()
		config.DB.Model(&key).Update("last_used", now)

		c.Set("user_id", key.UserID)
		c.Set("is_adminer", user.IsAdminer)
		c.Next()
	}
}

// OptionalAuthOrApiKey 可选认证：优先尝试 JWT，其次 API Key，均不存在则放行
func OptionalAuthOrApiKey(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT 认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err == nil && token.Valid {
				claims, ok := token.Claims.(jwt.MapClaims)
				if ok {
					userID := uint(claims["user_id"].(float64))
					isAdmin, _ := claims["is_adminer"].(bool)
					var user model.User
					if err := config.DB.First(&user, userID).Error; err == nil && user.Status != 0 {
						c.Set("user_id", userID)
						c.Set("is_adminer", isAdmin)
					}
				}
			}
		}

		// API Key 认证
		apiKey := c.GetHeader("X-Api-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		if apiKey != "" {
			var key model.ApiKey
			if err := config.DB.Where("key = ? AND revoked_at IS NULL", apiKey).First(&key).Error; err == nil {
				var user model.User
				if err := config.DB.First(&user, key.UserID).Error; err == nil && user.Status != 0 {
					now := time.Now()
					config.DB.Model(&key).Update("last_used", now)
					c.Set("user_id", key.UserID)
					c.Set("is_adminer", user.IsAdminer)
				}
			}
		}

		c.Next()
	}
}

// AuthOrApiKey 同时支持 JWT Bearer Token 和 X-Api-Key / ?api_key=xxx 两种认证方式
// 优先尝试 JWT，如果 JWT 不存在则尝试 API Key
func AuthOrApiKey(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// JWT 认证
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err == nil && token.Valid {
				claims, ok := token.Claims.(jwt.MapClaims)
				if ok {
					userID := uint(claims["user_id"].(float64))
					isAdmin, _ := claims["is_adminer"].(bool)
					var user model.User
					if err := config.DB.First(&user, userID).Error; err == nil && user.Status != 0 {
						c.Set("user_id", userID)
						c.Set("is_adminer", isAdmin)
						c.Next()
						return
					}
				}
			}
		}

		// API Key 认证
		apiKey := c.GetHeader("X-Api-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		if apiKey == "" {
			model.Fail(c, http.StatusUnauthorized, "未认证，请提供 Authorization Bearer Token 或 X-Api-Key")
			c.Abort()
			return
		}

		var key model.ApiKey
		if err := config.DB.Where("key = ? AND revoked_at IS NULL", apiKey).First(&key).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "API Key 无效")
			c.Abort()
			return
		}

		var user model.User
		if err := config.DB.First(&user, key.UserID).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "用户不存在")
			c.Abort()
			return
		}
		if user.Status == 0 {
			model.Fail(c, http.StatusForbidden, "账号已被冻结")
			c.Abort()
			return
		}

		now := time.Now()
		config.DB.Model(&key).Update("last_used", now)

		c.Set("user_id", key.UserID)
		c.Set("is_adminer", user.IsAdminer)
		c.Next()
	}
}
