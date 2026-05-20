package handler

import (
	"net/http"
	"strconv"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type ApiKeyHandler struct{}

func NewApiKeyHandler() *ApiKeyHandler {
	return &ApiKeyHandler{}
}

// List 获取当前用户的 API Key 列表
func (h *ApiKeyHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	var keys []model.ApiKey
	config.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&keys)
	model.Success(c, "获取成功", keys)
}

// Create 创建新的 API Key
func (h *ApiKeyHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		model.Fail(c, http.StatusBadRequest, "请输入 Key 名称")
		return
	}

	// 限制每个用户最多 10 个有效 Key
	var count int64
	config.DB.Model(&model.ApiKey{}).Where("user_id = ? AND revoked_at IS NULL", userID).Count(&count)
	if count >= 10 {
		model.Fail(c, http.StatusBadRequest, "每个用户最多创建 10 个 API Key")
		return
	}

	key := model.ApiKey{
		UserID: userID,
		Name:   req.Name,
		Key:    "lsky-" + model.RandomString(32),
	}

	if err := config.DB.Create(&key).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}

	model.Success(c, "创建成功", key)
}

// Revoke 撤销 API Key
func (h *ApiKeyHandler) Revoke(c *gin.Context) {
	userID := c.GetUint("user_id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		model.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	var key model.ApiKey
	if err := config.DB.Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).First(&key).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "Key 不存在或已撤销")
		return
	}

	now := time.Now()
	config.DB.Model(&key).Update("revoked_at", now)
	model.Success(c, "已撤销", nil)
}
