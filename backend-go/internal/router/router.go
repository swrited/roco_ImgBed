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

	// 静态文件：背景图片
	r.Static("/uploads/bg", "./uploads/bg")

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
		// ======== 公开设置（无需任何认证） ========
		api.GET("/settings/public", handler.NewPublicSettingHandler().Index)

		apiPub := api.Group("")
		apiPub.Use(middleware.OptionalAuthOrApiKey(cfg))
		{
			imgH := handler.NewImageHandler()
			apiPub.GET("/gallery", imgH.Gallery)
			apiPub.GET("/gallery/albums/:id", imgH.GalleryAlbum)
			apiPub.GET("/strategies", handler.NewStrategyHandler().List)
		}

		apiKeyOnly := api.Group("")
		apiKeyOnly.Use(middleware.ApiKeyAuth())
		{
			imgH := handler.NewImageHandler()
			apiKeyOnly.GET("/images/random", imgH.Random)
			apiKeyOnly.GET("/images/adaptive", imgH.Adaptive)
		}

		// Private image short links carry a per-image access token.
		api.GET("/images/:access_token/:key", handler.NewImageHandler().ShortLinkContent)

		// A user-scoped read token only retrieves that user's image data.
		readTokenImages := api.Group("/t/:token")
		readTokenImages.Use(middleware.ImageReadTokenAuth())
		{
			imgH := handler.NewImageHandler()
			readTokenImages.GET("/images", imgH.ListImages)
			readTokenImages.GET("/images/random", imgH.Random)
			readTokenImages.GET("/images/adaptive", imgH.Adaptive)
		}

		// Upload supports guest, logged-in browser sessions, or API Key clients.
		uploadGroup := api.Group("")
		uploadGroup.Use(middleware.OptionalAuthOrApiKey(cfg))
		{
			uploadGroup.POST("/upload", handler.NewImageHandler().Upload)
		}

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
	tagH := handler.NewTagHandler()
	apiKeyH := handler.NewApiKeyHandler()
	apiUsageH := handler.NewAPIUsageHandler()
	aiImageH := handler.NewAIImageHandler()

	// Profile
	g.GET("/profile", userH.Profile)
	g.PUT("/profile", userH.UpdateProfile)
	g.GET("/image-read-token", userH.ImageReadToken)
	g.PUT("/image-read-token", userH.ResetImageReadToken)

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
	g.GET("/images/trash", imgH.TrashList)
	g.PUT("/images/trash/restore", imgH.RestoreTrash)
	g.DELETE("/images/trash/force", imgH.ForceDeleteTrash)
	g.DELETE("/images", imgH.BatchDelete)
	g.DELETE("/images/:key", imgH.Delete)
	g.PUT("/images/rename", imgH.Rename)
	g.PUT("/images/movement", imgH.Move)
	g.PUT("/images/permission", imgH.UpdatePermission)
	g.PUT("/images/:key/access-token", imgH.ResetAccessToken)
	g.PUT("/images/:key/tags", imgH.UpdateTags)

	// AI image generation
	g.POST("/ai/images", aiImageH.Generate)

	// Albums
	g.GET("/albums", albumH.List)
	g.POST("/albums", albumH.Create)
	g.PUT("/albums/:id", albumH.Update)
	g.DELETE("/albums/:id", albumH.Delete)

	// Tags
	g.GET("/tags", tagH.List)
	g.DELETE("/tags/:id", tagH.Delete)

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
		adminGroup.POST("/settings/bg-upload", adminSettingH.BgUpload)
	}
}
