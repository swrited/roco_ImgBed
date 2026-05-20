package admin

import (
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type StrategyHandler struct{}

func NewStrategyHandler() *StrategyHandler { return &StrategyHandler{} }

func (h *StrategyHandler) List(c *gin.Context) {
	var strategies []model.Strategy
	config.DB.Order("id ASC").Find(&strategies)
	model.Success(c, "success", strategies)
}

type StrategyInput struct {
	Key     uint              `json:"key" binding:"required"`
	Name    string            `json:"name" binding:"required,max=64"`
	Intro   string            `json:"intro"`
	Configs model.JSONMap     `json:"configs"`
}

func (h *StrategyHandler) Create(c *gin.Context) {
	var input StrategyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	// 校验策略类型
	if _, ok := model.StrategyKeyNames[input.Key]; !ok {
		model.Fail(c, http.StatusUnprocessableEntity, "不支持的存储类型")
		return
	}

	if input.Configs == nil {
		input.Configs = model.JSONMap{}
	}

	s := model.Strategy{
		Key:     input.Key,
		Name:    input.Name,
		Intro:   input.Intro,
		Configs: input.Configs,
	}
	if err := config.DB.Create(&s).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}

	model.Success(c, "创建成功", s)
}

type StrategyUpdateInput struct {
	Name    string        `json:"name" binding:"max=64"`
	Intro   string        `json:"intro"`
	Configs model.JSONMap `json:"configs"`
}

func (h *StrategyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input StrategyUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Intro != "" {
		updates["intro"] = input.Intro
	}
	if input.Configs != nil {
		updates["configs"] = input.Configs
	}

	if len(updates) == 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "没有需要更新的字段")
		return
	}

	config.DB.Model(&model.Strategy{}).Where("id = ?", id).Updates(updates)
	model.Success(c, "更新成功", nil)
}

func (h *StrategyHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// 清理关联的 group_strategy 记录，避免关联失效
	config.DB.Where("strategy_id = ?", id).Delete(&model.GroupStrategy{})

	if err := config.DB.Delete(&model.Strategy{}, id).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	model.Success(c, "删除成功", nil)
}
