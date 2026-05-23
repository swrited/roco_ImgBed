package middleware

import (
	"net/http"
	"strconv"
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
		c.Set("auth_type", "bearer")
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
		c.Set("auth_type", "bearer")
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

func APIUsageLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if !shouldLogAPIUsage(path) {
			return
		}

		var userID *uint
		if id := c.GetUint("user_id"); id > 0 {
			userID = &id
		}

		var apiKeyID *uint
		if id := c.GetUint("api_key_id"); id > 0 {
			apiKeyID = &id
		}

		authType := ""
		if v, ok := c.Get("auth_type"); ok {
			authType, _ = v.(string)
		}
		if authType == "" {
			authType = "anonymous"
		}
		if authType != "api_key" {
			return
		}

		userAgent := c.GetHeader("User-Agent")
		if len(userAgent) > 255 {
			userAgent = userAgent[:255]
		}

		config.DB.Create(&model.ApiUsageLog{
			UserID:    userID,
			ApiKeyID:  apiKeyID,
			Method:    c.Request.Method,
			Path:      path,
			Status:    c.Writer.Status(),
			LatencyMS: time.Since(start).Milliseconds(),
			IP:        c.ClientIP(),
			UserAgent: userAgent,
			AuthType:  authType,
		})
	}
}

func shouldLogAPIUsage(path string) bool {
	if path == "" {
		return false
	}
	if strings.HasPrefix(path, "/api/v1/api-usage") || strings.HasPrefix(path, "/api/v1/admin/api-usage") {
		return false
	}
	return strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/images/")
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

		c.Set("user_id", key.UserID)
		c.Set("is_adminer", user.IsAdminer)
		c.Set("api_key_id", key.ID)
		c.Set("auth_type", "api_key")
		if !checkAPIKeyLimits(c, key.ID) {
			return
		}

		// Update last_used
		now := time.Now()
		config.DB.Model(&key).Update("last_used", now)
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
						c.Set("auth_type", "bearer")
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
					c.Set("api_key_id", key.ID)
					c.Set("auth_type", "api_key")
					if !checkAPIKeyLimits(c, key.ID) {
						return
					}
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
						c.Set("auth_type", "bearer")
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

		c.Set("user_id", key.UserID)
		c.Set("is_adminer", user.IsAdminer)
		c.Set("api_key_id", key.ID)
		c.Set("auth_type", "api_key")
		if !checkAPIKeyLimits(c, key.ID) {
			return
		}

		now := time.Now()
		config.DB.Model(&key).Update("last_used", now)
		c.Next()
	}
}

func checkAPIKeyLimits(c *gin.Context, apiKeyID uint) bool {
	if apiKeyID == 0 {
		return true
	}
	now := time.Now()

	minuteLimit := systemConfigInt("api_key_minute_limit", 60)
	if minuteLimit > 0 {
		start := now.Add(-time.Minute)
		if countAPIKeyUsage(apiKeyID, start) >= int64(minuteLimit) {
			model.Fail(c, http.StatusTooManyRequests, "API Key 请求过于频繁，请稍后再试")
			c.Abort()
			return false
		}
	}

	dailyLimit := systemConfigInt("api_key_daily_limit", 1000)
	if dailyLimit > 0 {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if countAPIKeyUsage(apiKeyID, start) >= int64(dailyLimit) {
			model.Fail(c, http.StatusTooManyRequests, "今日 API Key 请求次数已用完")
			c.Abort()
			return false
		}
	}

	return true
}

func countAPIKeyUsage(apiKeyID uint, since time.Time) int64 {
	var count int64
	config.DB.Model(&model.ApiUsageLog{}).
		Where("auth_type = ? AND api_key_id = ? AND created_at >= ?", "api_key", apiKeyID, since).
		Count(&count)
	return count
}

func systemConfigInt(key string, fallback int) int {
	var cfg model.SystemConfig
	if err := config.DB.Where("name = ?", key).First(&cfg).Error; err != nil {
		return fallback
	}
	value := strings.TrimSpace(cfg.Value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// TokenPathAuth 允许通过 URL 路径参数 :token 传递 API Key 进行认证
func TokenPathAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			model.Fail(c, http.StatusUnauthorized, "缺少 token 参数")
			c.Abort()
			return
		}

		var key model.ApiKey
		if err := config.DB.Where("key = ? AND revoked_at IS NULL", token).First(&key).Error; err != nil {
			model.Fail(c, http.StatusUnauthorized, "Token 无效")
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

		c.Set("user_id", key.UserID)
		c.Set("is_adminer", user.IsAdminer)
		c.Set("api_key_id", key.ID)
		c.Set("auth_type", "api_key")
		if !checkAPIKeyLimits(c, key.ID) {
			return
		}

		// Update last_used
		now := time.Now()
		config.DB.Model(&key).Update("last_used", now)
		c.Next()
	}
}
