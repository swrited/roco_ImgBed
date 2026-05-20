package admin

import (
	"net/http"
	"strconv"
	"strings"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
	"lskypro-server/internal/service/storage"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct{}

func NewImageHandler() *ImageHandler { return &ImageHandler{} }

func (h *ImageHandler) List(c *gin.Context) {
	page := 1
	perPage := 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil && pp > 0 && pp <= 100 {
		perPage = pp
	}

	var images []model.Image
	var total int64

	query := config.DB.Model(&model.Image{})

	if q := c.Query("q"); q != "" {
		query = query.Where("origin_name LIKE ? OR alias_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query.Count(&total)
	query.Offset((page-1)*perPage).Limit(perPage).Order("id DESC").Preload("User").Find(&images)

	imageDTOs := buildImageDTOs(images)

	model.Success(c, "success", gin.H{
		"data":         imageDTOs,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	var input struct {
		Keys []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.Keys) == 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "请提供要删除的图片")
		return
	}

	var images []model.Image
	config.DB.Where("`key` IN ?", input.Keys).Find(&images)

	for _, img := range images {
		// 通过存储适配器删除物理文件
		if img.StrategyID != nil {
			var strategy model.Strategy
			if err := config.DB.First(&strategy, *img.StrategyID).Error; err == nil {
				if adapter, err := storage.Factory(&strategy); err == nil {
					adapter.Delete(img.Pathname())
				}
			}
		}

		if img.UserID != nil {
			config.DB.Model(&model.User{}).Where("id = ?", *img.UserID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))
		}
	}

	config.DB.Where("`key` IN ?", input.Keys).Delete(&model.Image{})
	model.Success(c, "删除成功", nil)
}

// imageDTO wraps model.Image with a computed URL field
type imageDTO struct {
	model.Image
	URL string `json:"url"`
}

// buildImageDTOs 为图片列表构建带 URL 的 DTO
func buildImageDTOs(images []model.Image) []imageDTO {
	if len(images) == 0 {
		return nil
	}

	// 收集唯一的策略 ID
	strategyIDs := make(map[uint]bool)
	for _, img := range images {
		if img.StrategyID != nil {
			strategyIDs[*img.StrategyID] = true
		}
	}

	// 批量加载策略并构建 URL 映射
	strategyURLs := make(map[uint]string)
	if len(strategyIDs) > 0 {
		ids := make([]uint, 0, len(strategyIDs))
		for id := range strategyIDs {
			ids = append(ids, id)
		}

		var strategies []model.Strategy
		config.DB.Where("id IN ?", ids).Find(&strategies)
		for _, s := range strategies {
			strategyURLs[s.ID] = storage.GetStrategyURL(&s)
		}
	}

	// 构建 DTO
	result := make([]imageDTO, len(images))
	for i, img := range images {
		dto := imageDTO{Image: img}
		if img.StrategyID != nil {
			if baseURL, ok := strategyURLs[*img.StrategyID]; ok {
				if !strings.HasSuffix(baseURL, "/") {
					baseURL += "/"
				}
				dto.URL = baseURL + img.Pathname()
			}
		}
		result[i] = dto
	}
	return result
}
