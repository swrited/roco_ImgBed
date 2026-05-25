package handler

import (
	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

// PublicSettingHandler 提供无需认证的公开系统配置
type PublicSettingHandler struct{}

func NewPublicSettingHandler() *PublicSettingHandler { return &PublicSettingHandler{} }

// 白名单：只返回前端展示所需的公开配置项
var publicSettingKeys = []string{
	"app_name",
	"app_version",
	"site_description",
	"site_keywords",
	"is_enable_registration",
	"is_enable_gallery",
	"is_enable_api",
	"is_enable_ai_image",
	"ai_image_provider",
	"site_bg_image",
	"site_bg_opacity",
}

func (h *PublicSettingHandler) Index(c *gin.Context) {
	var settings []model.SystemConfig
	config.DB.Where("name IN ?", publicSettingKeys).Find(&settings)

	data := make(gin.H)
	for _, s := range settings {
		data[s.Name] = s.Value
	}

	model.Success(c, "success", data)
}
