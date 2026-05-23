package handler

import (
	"strconv"
	"strings"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type APIUsageHandler struct{}

func NewAPIUsageHandler() *APIUsageHandler { return &APIUsageHandler{} }

type APIUsageDayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type APIUsageEndpointStat struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Count  int64  `json:"count"`
}

type APIUsageUserStat struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Count  int64  `json:"count"`
}

type APIUsageLogRow struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	ApiKeyID   uint      `json:"api_key_id"`
	ApiKeyName string    `json:"api_key_name"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	LatencyMS  int64     `json:"latency_ms"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *APIUsageHandler) UserStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	start, end := usageRange(c)
	base := config.DB.Model(&model.ApiUsageLog{}).Where("auth_type = ? AND user_id = ?", "api_key", userID)
	model.Success(c, "success", usageStats(c, base, false, start, end))
}

func (h *APIUsageHandler) AdminStats(c *gin.Context) {
	start, end := usageRange(c)
	base := config.DB.Model(&model.ApiUsageLog{}).Where("auth_type = ?", "api_key")
	data := usageStats(c, base, true, start, end)

	var users []APIUsageUserStat
	config.DB.Table("api_usage_logs").
		Select("users.id as user_id, users.name, users.email, COUNT(*) as count").
		Joins("LEFT JOIN users ON users.id = api_usage_logs.user_id").
		Where("api_usage_logs.auth_type = ? AND api_usage_logs.user_id IS NOT NULL AND api_usage_logs.created_at >= ? AND api_usage_logs.created_at < ?", "api_key", start, end).
		Group("users.id, users.name, users.email").
		Order("count DESC").
		Limit(10).
		Scan(&users)
	data["by_user"] = users

	model.Success(c, "success", data)
}

func usageStats(c *gin.Context, base *gorm.DB, includeRecent bool, start time.Time, end time.Time) gin.H {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -6)
	rangeBase := base.Session(&gorm.Session{}).Where("created_at >= ? AND created_at < ?", start, end)

	var total int64
	rangeBase.Count(&total)

	var todayCount int64
	base.Session(&gorm.Session{}).Where("created_at >= ?", today).Count(&todayCount)

	var weekCount int64
	base.Session(&gorm.Session{}).Where("created_at >= ?", weekAgo).Count(&weekCount)

	var days []APIUsageDayStat
	base.Session(&gorm.Session{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", weekAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&days)

	var endpoints []APIUsageEndpointStat
	rangeBase.Session(&gorm.Session{}).
		Select("method, path, COUNT(*) as count").
		Group("method, path").
		Order("count DESC").
		Limit(10).
		Scan(&endpoints)
	if endpoints == nil {
		endpoints = []APIUsageEndpointStat{}
	}

	data := gin.H{
		"total":       total,
		"today":       todayCount,
		"last_7_days": weekCount,
		"daily":       fillUsageDays(days, weekAgo),
		"endpoints":   endpoints,
		"range": gin.H{
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
		},
	}

	if includeRecent {
		limit := usageLimit(c)
		var recent []APIUsageLogRow
		config.DB.Table("api_usage_logs").
			Select("api_usage_logs.id, api_usage_logs.user_id, users.name, users.email, api_usage_logs.api_key_id, api_keys.name as api_key_name, api_usage_logs.method, api_usage_logs.path, api_usage_logs.status, api_usage_logs.latency_ms, api_usage_logs.ip, api_usage_logs.user_agent, api_usage_logs.created_at").
			Joins("LEFT JOIN users ON users.id = api_usage_logs.user_id").
			Joins("LEFT JOIN api_keys ON api_keys.id = api_usage_logs.api_key_id").
			Where("api_usage_logs.auth_type = ? AND api_usage_logs.created_at >= ? AND api_usage_logs.created_at < ?", "api_key", start, end).
			Order("api_usage_logs.id DESC").
			Limit(limit).
			Scan(&recent)
		data["recent"] = recent
	}

	return data
}

func fillUsageDays(rows []APIUsageDayStat, start time.Time) []APIUsageDayStat {
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Date] = row.Count
	}
	days := make([]APIUsageDayStat, 0, 7)
	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		days = append(days, APIUsageDayStat{Date: date, Count: counts[date]})
	}
	return days
}

func usageRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	if v := strings.TrimSpace(c.Query("start")); v != "" {
		if t, ok := parseUsageTime(v, false); ok {
			start = t
		}
	}
	if v := strings.TrimSpace(c.Query("end")); v != "" {
		if t, ok := parseUsageTime(v, true); ok {
			end = t
		}
	}
	if !end.After(start) {
		end = start.AddDate(0, 0, 1)
	}
	return start, end
}

func parseUsageTime(value string, endOfDay bool) (time.Time, bool) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			if layout == "2006-01-02" && endOfDay {
				return t.AddDate(0, 0, 1), true
			}
			return t, true
		}
	}
	return time.Time{}, false
}

func usageLimit(c *gin.Context) int {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if err != nil || limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}
