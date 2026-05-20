package handler

import (
	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type StrategyHandler struct{}

func NewStrategyHandler() *StrategyHandler { return &StrategyHandler{} }

func (h *StrategyHandler) List(c *gin.Context) {
	var strategies []model.Strategy
	config.DB.Order("id ASC").Find(&strategies)

	output := make([]gin.H, len(strategies))
	for i, s := range strategies {
		output[i] = gin.H{
			"id":      s.ID,
			"key":     s.Key,
			"name":    s.Name,
			"intro":   s.Intro,
		}
	}

	model.Success(c, "success", output)
}
