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
	r.Use(middleware.APIUsageLogger())

	// Health check
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Public random image aliases for direct backend usage.
	publicImageH := handler.NewImageHandler()
	publicImages := r.Group("/images")
	publicImages.Use(middleware.ApiKeyAuth())
	{
		publicImages.GET("/random", publicImageH.Random)
		publicImages.GET("/adaptive", publicImageH.Adaptive)
	}

	api := r.Group("/api/v1")
	{
		authH := handler.NewAuthHandler(cfg)
		// ======== 完全公开路由（无需任何认证） ========
		api.POST("/tokens", authH.Login)
		api.POST("/register", authH.Register)
		api.POST("/forgot-password", handler.NewUserHandler().ForgotPassword)
		api.POST("/reset-password", handler.NewUserHandler().ResetPassword)

		// ======== 公开 API（支持 Bearer Token、API Key 或无认证） ========
		apiPub := api.Group("")
		apiPub.Use(middleware.OptionalAuthOrApiKey(cfg))
		{
			imgH := handler.NewImageHandler()
			apiPub.GET("/gallery", imgH.Gallery)
			apiPub.GET("/strategies", handler.NewStrategyHandler().List)
		}

		apiKeyOnly := api.Group("")
		apiKeyOnly.Use(middleware.ApiKeyAuth())
		{
			imgH := handler.NewImageHandler()
			apiKeyOnly.GET("/images/random", imgH.Random)
			apiKeyOnly.GET("/images/adaptive", imgH.Adaptive)
		}

		// Upload (optional auth - supports guest, Bearer token, or api_key)
		uploadGroup := api.Group("")
		uploadGroup.Use(middleware.OptionalAuthOrApiKey(cfg))
		{
			uploadGroup.POST("/upload", handler.NewImageHandler().Upload)
		}

		// ======== 无认证 URL 路径参数路由（支持所有业务接口） ========
		tokenAuthed := api.Group("/t/:token")
		tokenAuthed.Use(middleware.TokenPathAuth())
		// 单独为 upload 和 random/adaptive 加上，因为它们不在常规 authed 组里
		tokenAuthed.POST("/upload", handler.NewImageHandler().Upload)
		tokenAuthed.GET("/images/random", handler.NewImageHandler().Random)
		tokenAuthed.GET("/images/adaptive", handler.NewImageHandler().Adaptive)
		registerAuthedRoutes(tokenAuthed, authH)

		// ======== 认证路由（支持 Bearer Token 或 API Key） ========
		authed := api.Group("")
		authed.Use(middleware.AuthOrApiKey(cfg))
		registerAuthedRoutes(authed, authH)


	}

	return r
}

func registerAuthedRoutes(g *gin.RouterGroup, authH *handler.AuthHandler) {
	userH := handler.NewUserHandler()
	imgH := handler.NewImageHandler()
	albumH := handler.NewAlbumHandler()
	apiKeyH := handler.NewApiKeyHandler()
	apiUsageH := handler.NewAPIUsageHandler()
	aiImageH := handler.NewAIImageHandler()

	// Profile
	g.GET("/profile", userH.Profile)
	g.PUT("/profile", userH.UpdateProfile)

	// Dashboard
	g.GET("/dashboard", userH.Dashboard)

	// User settings
	g.GET("/user/settings", userH.Settings)
	g.PUT("/user/settings", userH.UpdateSettings)
	g.PUT("/user/settings/strategy", userH.SetStrategy)
	g.PUT("/user/settings/permission", userH.SetPermission)
	g.PUT("/user/settings/album", userH.SetAlbum)

	// API Keys management
	g.GET("/api-keys", apiKeyH.List)
	g.POST("/api-keys", apiKeyH.Create)
	g.DELETE("/api-keys/:id", apiKeyH.Revoke)
	g.GET("/api-usage", apiUsageH.UserStats)

	// Images
	g.GET("/images", imgH.ListImages)
	g.DELETE("/images", imgH.BatchDelete)
	g.DELETE("/images/:key", imgH.Delete)
	g.PUT("/images/rename", imgH.Rename)
	g.PUT("/images/movement", imgH.Move)
	g.PUT("/images/permission", imgH.Permission)

	// AI image generation
	g.POST("/ai/images", aiImageH.Generate)

	// Albums
	g.GET("/albums", albumH.List)
	g.POST("/albums", albumH.Create)
	g.PUT("/albums/:id", albumH.Update)
	g.DELETE("/albums/:id", albumH.Delete)

	// Token management
	g.DELETE("/tokens", authH.Logout)

	// Admin routes
	adminGroup := g.Group("/admin")
	adminGroup.Use(middleware.AdminRequired())
	{
		consoleH := admin.NewConsoleHandler()
		adminUserH := admin.NewUserHandler()
		adminImageH := admin.NewImageHandler()
		adminGroupH := admin.NewGroupHandler()
		adminStrategyH := admin.NewStrategyHandler()
		adminSettingH := admin.NewSettingHandler()
		adminAPIUsageH := handler.NewAPIUsageHandler()

		adminGroup.GET("/console", consoleH.Index)
		adminGroup.GET("/api-usage", adminAPIUsageH.AdminStats)

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
