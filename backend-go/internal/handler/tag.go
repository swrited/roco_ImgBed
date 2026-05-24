package handler

import (
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type TagHandler struct{}

func NewTagHandler() *TagHandler { return &TagHandler{} }

func (h *TagHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	var tags []model.Tag
	config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&tags)

	var result []map[string]interface{}
	for _, t := range tags {
		var count int64
		config.DB.Table("image_tags").Where("tag_id = ?", t.ID).Count(&count)
		result = append(result, map[string]interface{}{
			"id":         t.ID,
			"name":       t.Name,
			"image_num":  count,
			"created_at": t.CreatedAt,
		})
	}

	model.Success(c, "success", result)
}

func (h *TagHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	tagID := c.Param("id")

	var tag model.Tag
	if err := config.DB.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "标签不存在")
		return
	}

	// 1. Delete relations
	config.DB.Exec("DELETE FROM image_tags WHERE tag_id = ?", tag.ID)
	// 2. Delete tag
	config.DB.Delete(&tag)

	model.Success(c, "删除成功", nil)
}
