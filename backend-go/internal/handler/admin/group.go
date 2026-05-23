package admin

import (
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct{}

func NewGroupHandler() *GroupHandler { return &GroupHandler{} }

type GroupListItem struct {
	model.Group
	UsersCount      int64 `json:"users_count"`
	StrategiesCount int64 `json:"strategies_count"`
}

func (h *GroupHandler) List(c *gin.Context) {
	var groups []model.Group
	config.DB.Preload("Strategies").Order("id ASC").Find(&groups)

	items := make([]GroupListItem, 0, len(groups))
	for _, group := range groups {
		var usersCount int64
		config.DB.Model(&model.User{}).Where("group_id = ?", group.ID).Count(&usersCount)
		items = append(items, GroupListItem{
			Group:           group,
			UsersCount:      usersCount,
			StrategiesCount: int64(len(group.Strategies)),
		})
	}

	model.Success(c, "success", items)
}

type GroupInput struct {
	Name      string        `json:"name" binding:"required,max=64"`
	IsDefault bool          `json:"is_default"`
	IsGuest   bool          `json:"is_guest"`
	Configs   model.JSONMap `json:"configs"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var input GroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	group := model.Group{
		Name:      input.Name,
		IsDefault: input.IsDefault,
		IsGuest:   input.IsGuest,
		Configs:   normalizeGroupConfigs(input.Configs),
	}
	if group.IsDefault {
		config.DB.Model(&model.Group{}).Update("is_default", false)
	}
	if group.IsGuest {
		config.DB.Model(&model.Group{}).Update("is_guest", false)
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
		if boolValue(v) {
			config.DB.Model(&model.Group{}).Where("id != ?", id).Update("is_default", false)
		}
		updates["is_default"] = v
	}
	if v, ok := input["is_guest"]; ok {
		if boolValue(v) {
			config.DB.Model(&model.Group{}).Where("id != ?", id).Update("is_guest", false)
		}
		updates["is_guest"] = v
	}
	if v, ok := input["configs"]; ok {
		if cfg, ok := v.(map[string]interface{}); ok {
			updates["configs"] = normalizeGroupConfigs(model.JSONMap(cfg))
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
	}

	model.Success(c, "更新成功", nil)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var group model.Group
	if err := config.DB.Where("id = ?", id).First(&group).Error; err == nil {
		if group.IsDefault || group.IsGuest {
			model.Fail(c, http.StatusBadRequest, "默认组和游客组无法删除")
			return
		}
		var defaultGroup model.Group
		if err := config.DB.Where("is_default = ?", true).First(&defaultGroup).Error; err == nil {
			config.DB.Model(&model.User{}).Where("group_id = ?", group.ID).Update("group_id", defaultGroup.ID)
		}
	}
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

func boolValue(value interface{}) bool {
	v, ok := value.(bool)
	return ok && v
}

func normalizeGroupConfigs(configs model.JSONMap) model.JSONMap {
	defaults := model.JSONMap{
		"maximum_file_size":             5120,
		"concurrent_upload_num":         3,
		"is_enable_scan":                false,
		"is_enable_watermark":           false,
		"is_enable_original_protection": false,
		"scanned_action":                "mark",
		"scan_configs":                  model.JSONMap{"driver": "tencent"},
		"watermark_configs":             model.JSONMap{"mode": 1, "driver": "font"},
		"limit_per_minute":              20,
		"limit_per_hour":                100,
		"limit_per_day":                 300,
		"limit_per_week":                600,
		"limit_per_month":               999,
		"accepted_file_suffixes":        []string{"jpeg", "jpg", "png", "gif", "tif", "bmp", "ico", "psd", "webp"},
		"image_save_format":             "",
		"image_save_quality":            100,
		"path_naming_rule":              "{Y}/{m}/{d}",
		"file_naming_rule":              "{uniqid}",
		"image_cache_ttl":               2626560,
	}
	if configs == nil {
		return defaults
	}
	for key, value := range configs {
		defaults[key] = value
	}
	return defaults
}
