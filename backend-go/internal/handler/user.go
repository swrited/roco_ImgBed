package handler

import (
	"net/http"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler { return &UserHandler{} }

func (h *UserHandler) Profile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	var usedCapacity float64
	config.DB.Model(&model.Image{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&usedCapacity)

	model.Success(c, "success", gin.H{
		"id":            user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"is_adminer":    user.IsAdminer,
		"capacity":      user.Capacity,
		"used_capacity": usedCapacity,
		"image_num":     user.ImageNum,
		"album_num":     user.AlbumNum,
		"url":           user.URL,
		"avatar":        user.ToProfile().Avatar,
		"token":         user.Token,
		"status":        user.Status,
		"group_id":      user.GroupID,
		"configs":       user.Configs,
		"created_at":    user.CreatedAt,
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	newToken := model.RandomString(6)
	if err := config.DB.Model(&user).Update("token", newToken).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "刷新失败")
		return
	}

	model.Success(c, "更新成功", gin.H{"token": newToken})
}

type UpdateProfileInput struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	userID := c.GetUint("user_id")
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.URL != "" {
		updates["url"] = input.URL
	}
	if input.Password != "" {
		if len(input.Password) < 6 {
			model.Fail(c, http.StatusUnprocessableEntity, "密码至少6位")
			return
		}
		if !config.CheckPassword(input.OldPassword, user.Password) {
			model.Fail(c, http.StatusUnprocessableEntity, "当前密码错误")
			return
		}
		hash, _ := config.HashPassword(input.Password)
		updates["password"] = hash
		updates["remember_token"] = model.RandomString(60)
	}

	if len(updates) > 0 {
		config.DB.Model(&user).Updates(updates)
	}

	model.Success(c, "更新成功", user.ToProfile())
}

func (h *UserHandler) Dashboard(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	config.DB.Preload("Group.Strategies").First(&user, userID)

	var imageCount int64
	config.DB.Model(&model.Image{}).Where("user_id = ?", userID).Count(&imageCount)

	var albumCount int64
	config.DB.Model(&model.Album{}).Where("user_id = ?", userID).Count(&albumCount)

	var usedCapacity float64
	config.DB.Model(&model.Image{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&usedCapacity)

	// 今日上传数
	var todayCount int64
	config.DB.Model(&model.Image{}).Where("user_id = ? AND DATE(created_at) = DATE('now')", userID).Count(&todayCount)

	// 本月上传数
	var monthCount int64
	config.DB.Model(&model.Image{}).Where("user_id = ? AND strftime('%Y-%m', created_at) = strftime('%Y-%m', 'now')", userID).Count(&monthCount)

	// 近30天每日上传统计
	type DayStat struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	dailyStats := make([]DayStat, 30)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -29)
	rows, err := config.DB.Model(&model.Image{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Rows()
	if err == nil {
		defer rows.Close()
		type row struct {
			Date  string
			Count int64
		}
		dbStats := make(map[string]int64)
		for rows.Next() {
			var r row
			if err := config.DB.ScanRows(rows, &r); err != nil {
				continue
			}
			dbStats[r.Date] = r.Count
		}
		for i := 0; i < 30; i++ {
			dateStr := thirtyDaysAgo.AddDate(0, 0, i).Format("2006-01-02")
			dailyStats[i] = DayStat{Date: dateStr, Count: dbStats[dateStr]}
		}
	}

	model.Success(c, "success", gin.H{
		"user":          user.ToProfile(),
		"image_count":   imageCount,
		"album_count":   albumCount,
		"used_capacity": usedCapacity,
		"today_count":   todayCount,
		"month_count":   monthCount,
		"daily_stats":   dailyStats,
	})
}

func (h *UserHandler) Settings(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	model.Success(c, "success", gin.H{
		"name":            user.Name,
		"email":           user.Email,
		"url":             user.URL,
		"configs":         user.Configs,
		"upload_max_size": systemConfigInt("upload_max_size", 10240),
	})
}

func (h *UserHandler) UpdateSettings(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if v, ok := input["name"]; ok {
		updates["name"] = v
	}
	if v, ok := input["url"]; ok {
		updates["url"] = v
	}
	if v, ok := input["configs"]; ok {
		updates["configs"] = v
	}
	if v, ok := input["password"]; ok && v.(string) != "" {
		hash, _ := config.HashPassword(v.(string))
		updates["password"] = hash
	}

	if len(updates) > 0 {
		config.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
	}

	model.Success(c, "保存成功", nil)
}

func (h *UserHandler) SetStrategy(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		StrategyID uint `json:"strategy_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var user model.User
	config.DB.First(&user, userID)
	if user.Configs == nil {
		user.Configs = model.JSONMap{}
	}
	user.Configs["default_strategy"] = input.StrategyID
	config.DB.Model(&user).Update("configs", user.Configs)

	model.Success(c, "设置成功", nil)
}

func (h *UserHandler) SetPermission(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Permission uint `json:"permission"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var user model.User
	config.DB.First(&user, userID)
	if user.Configs == nil {
		user.Configs = model.JSONMap{}
	}
	user.Configs["default_permission"] = input.Permission
	config.DB.Model(&user).Update("configs", user.Configs)

	model.Success(c, "设置成功", nil)
}

func (h *UserHandler) SetAlbum(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		AlbumID *uint `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	if user.Configs == nil {
		user.Configs = model.JSONMap{}
	}

	if input.AlbumID == nil || *input.AlbumID == 0 {
		delete(user.Configs, "default_album_id")
	} else {
		var album model.Album
		if err := config.DB.Where("id = ? AND user_id = ?", *input.AlbumID, userID).First(&album).Error; err != nil {
			model.Fail(c, http.StatusUnprocessableEntity, "相册不存在")
			return
		}
		user.Configs["default_album_id"] = *input.AlbumID
	}

	config.DB.Model(&user).Update("configs", user.Configs)
	model.Success(c, "设置成功", nil)
}

// ForgotPassword sends a password reset email
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var user model.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Don't reveal whether user exists
		model.Success(c, "如果该邮箱已注册，重置邮件已发送", nil)
		return
	}

	token := model.RandomHex(32)
	config.DB.Create(&model.PasswordReset{
		Email:     input.Email,
		Token:     token,
		CreatedAt: time.Now(),
	})

	// In production, send email here
	// For now just succeed
	model.Success(c, "如果该邮箱已注册，重置邮件已发送", nil)
}

type ResetPasswordInput struct {
	Token    string `json:"token" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var reset model.PasswordReset
	if err := config.DB.Where("email = ? AND token = ?", input.Email, input.Token).First(&reset).Error; err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "重置令牌无效")
		return
	}

	// Check expiry (60 minutes)
	if time.Since(reset.CreatedAt) > 60*time.Minute {
		model.Fail(c, http.StatusUnprocessableEntity, "重置令牌已过期")
		return
	}

	hash, _ := config.HashPassword(input.Password)
	config.DB.Model(&model.User{}).Where("email = ?", input.Email).Updates(map[string]interface{}{
		"password":       hash,
		"remember_token": model.RandomString(60),
	})

	config.DB.Where("email = ?", input.Email).Delete(&model.PasswordReset{})

	model.Success(c, "密码重置成功", nil)
}
