package admin

import (
	"net/http"
	"strconv"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

type UserWithUsage struct {
	model.User
	UsedCapacity      float64 `json:"used_capacity"`
	RemainingCapacity float64 `json:"remaining_capacity"`
	CapacityPercent   int     `json:"capacity_percent"`
}

func (h *UserHandler) List(c *gin.Context) {
	page := 1
	perPage := 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}

	var users []model.User
	var total int64

	query := config.DB.Model(&model.User{})

	if q := c.Query("q"); q != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * perPage).Limit(perPage).Order("id DESC").Find(&users)
	data := attachUsage(users)

	model.Success(c, "success", gin.H{
		"data":         data,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *UserHandler) Edit(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	model.Success(c, "success", attachUsage([]model.User{user})[0])
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	for _, field := range []string{"name", "email", "url", "status"} {
		if v, ok := input[field]; ok {
			updates[field] = v
		}
	}
	if v, ok := input["capacity"]; ok {
		updates["capacity"] = v
	}
	if v, ok := input["is_adminer"]; ok {
		updates["is_adminer"] = v
	}
	if v, ok := input["password"]; ok && v.(string) != "" {
		hash, _ := config.HashPassword(v.(string))
		updates["password"] = hash
	}

	if len(updates) > 0 {
		config.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	}

	model.Success(c, "更新成功", nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&model.User{}, id).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	model.Success(c, "删除成功", nil)
}

func attachUsage(users []model.User) []UserWithUsage {
	data := make([]UserWithUsage, 0, len(users))
	if len(users) == 0 {
		return data
	}

	ids := make([]uint, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}

	type usageRow struct {
		UserID uint
		Used   float64
	}
	var rows []usageRow
	config.DB.Model(&model.Image{}).
		Select("user_id, COALESCE(SUM(size), 0) as used").
		Where("user_id IN ?", ids).
		Group("user_id").
		Scan(&rows)

	usageByUser := make(map[uint]float64, len(rows))
	for _, row := range rows {
		usageByUser[row.UserID] = row.Used
	}

	for _, user := range users {
		used := usageByUser[user.ID]
		remaining := float64(0)
		percent := 0
		if user.Capacity > 0 {
			remaining = user.Capacity - used
			if remaining < 0 {
				remaining = 0
			}
			percent = int((used / user.Capacity) * 100)
			if percent > 100 {
				percent = 100
			}
		}
		data = append(data, UserWithUsage{
			User:              user,
			UsedCapacity:      used,
			RemainingCapacity: remaining,
			CapacityPercent:   percent,
		})
	}

	return data
}
