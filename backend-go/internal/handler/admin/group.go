package admin

import (
	"fmt"
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct{}

func NewGroupHandler() *GroupHandler { return &GroupHandler{} }

func (h *GroupHandler) List(c *gin.Context) {
	var groups []model.Group
	config.DB.Preload("Strategies").Order("id ASC").Find(&groups)
	model.Success(c, "success", groups)
}

type GroupInput struct {
	Name string `json:"name" binding:"required,max=64"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var input GroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	group := model.Group{
		Name:    input.Name,
		Configs: model.JSONMap{},
	}
	if err := config.DB.Create(&group).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}

	model.Success(c, "创建成功", group)
}

func (h *GroupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var group model.Group
	if err := config.DB.Where("id = ?", id).First(&group).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "角色组不存在")
		return
	}

	updates := map[string]interface{}{}
	if v, ok := input["name"]; ok {
		updates["name"] = v
	}
	if v, ok := input["is_default"]; ok {
		if v.(bool) {
			config.DB.Model(&model.Group{}).Where("id != ?", id).Update("is_default", false)
		}
		updates["is_default"] = v
	}
	if v, ok := input["is_guest"]; ok {
		if v.(bool) {
			config.DB.Model(&model.Group{}).Where("id != ?", id).Update("is_guest", false)
		}
		updates["is_guest"] = v
	}
	if v, ok := input["configs"]; ok {
		if cfg, ok := v.(map[string]interface{}); ok {
			updates["configs"] = model.JSONMap(cfg)
		}
	}

	if len(updates) > 0 {
		config.DB.Model(&group).Updates(updates)
	}

	// Handle strategy association if provided
	if v, ok := input["strategy_ids"]; ok {
		config.DB.Where("group_id = ?", group.ID).Delete(&model.GroupStrategy{})
		if ids, ok := v.([]interface{}); ok && len(ids) > 0 {
			for _, sid := range ids {
				sid := uint(sid.(float64))
				config.DB.Create(&model.GroupStrategy{GroupID: group.ID, StrategyID: sid})
			}
		}
		fmt.Printf("DEBUG: [Group Update] group_id=%d strategy_ids=%v\n", group.ID, v)
	}

	model.Success(c, "更新成功", nil)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	config.DB.Where("group_id = ?", id).Delete(&model.GroupStrategy{})
	if err := config.DB.Delete(&model.Group{}, id).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	model.Success(c, "删除成功", nil)
}

func (h *GroupHandler) ClearCache(c *gin.Context) {
	model.Success(c, "缓存已清除", nil)
}
