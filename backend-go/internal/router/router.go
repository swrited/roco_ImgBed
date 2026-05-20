package router

import (
	"lskypro-server/internal/config"
	"lskypro-server/internal/handler"
	"lskypro-server/internal/handler/admin"
	"lskypro-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api/v1")
	{
		// ======== 完全公开路由（无需任何认证） ========
		authH := handler.NewAuthHandler(cfg)
		api.POST("/tokens", authH.Login)
		api.POST("/register", authH.Register)
		api.POST("/forgot-password", handler.NewUserHandler().ForgotPassword)
		api.POST("/reset-password", handler.NewUserHandler().ResetPassword)

		// ======== 需要 API Key 认证的公开 API ========
		apiPub := api.Group("")
		apiPub.Use(middleware.ApiKeyAuth())
		{
			apiPub.GET("/gallery", handler.NewImageHandler().Gallery)
			apiPub.GET("/strategies", handler.NewStrategyHandler().List)
		}

		// Upload (optional auth - supports guest, Bearer token, or api_key)
		uploadGroup := api.Group("")
		uploadGroup.Use(middleware.OptionalAuthOrApiKey(cfg))
		{
			uploadGroup.POST("/upload", handler.NewImageHandler().Upload)
		}

		// ======== 认证路由（支持 Bearer Token 或 API Key） ========
		authed := api.Group("")
		authed.Use(middleware.AuthOrApiKey(cfg))
		{
			userH := handler.NewUserHandler()
			imgH := handler.NewImageHandler()
			albumH := handler.NewAlbumHandler()
			apiKeyH := handler.NewApiKeyHandler()

			// Profile
			authed.GET("/profile", userH.Profile)
			authed.PUT("/profile", userH.UpdateProfile)

			// Dashboard
			authed.GET("/dashboard", userH.Dashboard)

			// User settings
			authed.GET("/user/settings", userH.Settings)
			authed.PUT("/user/settings", userH.UpdateSettings)
			authed.PUT("/user/settings/strategy", userH.SetStrategy)

			// API Keys management
			authed.GET("/api-keys", apiKeyH.List)
			authed.POST("/api-keys", apiKeyH.Create)
			authed.DELETE("/api-keys/:id", apiKeyH.Revoke)

			// Images
			authed.GET("/images", imgH.ListImages)
			authed.DELETE("/images", imgH.BatchDelete)
			authed.DELETE("/images/:key", imgH.Delete)
			authed.PUT("/images/rename", imgH.Rename)
			authed.PUT("/images/movement", imgH.Move)
			authed.PUT("/images/permission", imgH.Permission)

			// Albums
			authed.GET("/albums", albumH.List)
			authed.POST("/albums", albumH.Create)
			authed.PUT("/albums/:id", albumH.Update)
			authed.DELETE("/albums/:id", albumH.Delete)

			// Token management
			authed.DELETE("/tokens", authH.Logout)

			// Admin routes (AdminRequired 叠加在 AuthOrApiKey 之上)
			adminGroup := authed.Group("/admin")
			adminGroup.Use(middleware.AdminRequired())
			{
				consoleH := admin.NewConsoleHandler()
				adminUserH := admin.NewUserHandler()
				adminImageH := admin.NewImageHandler()
				adminGroupH := admin.NewGroupHandler()
				adminStrategyH := admin.NewStrategyHandler()
				adminSettingH := admin.NewSettingHandler()

				adminGroup.GET("/console", consoleH.Index)

				// Users
				adminGroup.GET("/users", adminUserH.List)
				adminGroup.GET("/users/:id", adminUserH.Edit)
				adminGroup.PUT("/users/:id", adminUserH.Update)
				adminGroup.DELETE("/users/:id", adminUserH.Delete)

				// Images
				adminGroup.GET("/images", adminImageH.List)
				adminGroup.DELETE("/images", adminImageH.Delete)

				// Groups
				adminGroup.GET("/groups", adminGroupH.List)
				adminGroup.POST("/groups", adminGroupH.Create)
				adminGroup.DELETE("/groups/clear-cache", adminGroupH.ClearCache)
				adminGroup.PUT("/groups/:id", adminGroupH.Update)
				adminGroup.DELETE("/groups/:id", adminGroupH.Delete)

				// Strategies
				adminGroup.GET("/strategies", adminStrategyH.List)
				adminGroup.POST("/strategies", adminStrategyH.Create)
				adminGroup.PUT("/strategies/:id", adminStrategyH.Update)
				adminGroup.DELETE("/strategies/:id", adminStrategyH.Delete)

				// Settings
				adminGroup.GET("/settings", adminSettingH.Index)
				adminGroup.PUT("/settings", adminSettingH.Save)
				adminGroup.POST("/settings/mail-test", adminSettingH.MailTest)
			}
		}
	}

	return r
}
