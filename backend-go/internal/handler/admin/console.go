package admin

import (
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type ConsoleHandler struct{}

func NewConsoleHandler() *ConsoleHandler { return &ConsoleHandler{} }

func (h *ConsoleHandler) Index(c *gin.Context) {
	var userCount, imageCount, albumCount int64
	config.DB.Model(&model.User{}).Count(&userCount)
	config.DB.Model(&model.Image{}).Count(&imageCount)
	config.DB.Model(&model.Album{}).Count(&albumCount)

	// 30-day stats
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	var dailyData []gin.H
	for i := 29; i >= 0; i-- {
		day := time.Now().Add(-time.Duration(i) * 24 * time.Hour)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
		end := start.Add(24 * time.Hour)

		var count int64
		config.DB.Model(&model.Image{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count)

		dailyData = append(dailyData, gin.H{
			"date":  start.Format("01-02"),
			"count": count,
		})
	}

	// Recent images
	var recentCount int64
	config.DB.Model(&model.Image{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&recentCount)

	model.Success(c, "success", gin.H{
		"stats": gin.H{
			"users":          userCount,
			"images":         imageCount,
			"albums":         albumCount,
			"recent_uploads": recentCount,
		},
		"daily": dailyData,
	})
}
