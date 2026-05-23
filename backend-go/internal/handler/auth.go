package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct{ cfg *config.Config }

func NewAuthHandler(cfg *config.Config) *AuthHandler { return &AuthHandler{cfg} }

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var user model.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		model.Fail(c, http.StatusUnauthorized, "账号不存在")
		return
	}

	if !config.CheckPassword(input.Password, user.Password) {
		model.Fail(c, http.StatusUnauthorized, "账号或密码错误")
		return
	}

	if user.Status == 0 {
		model.Fail(c, http.StatusForbidden, "账号已被冻结")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"is_adminer": user.IsAdminer,
		"exp":        time.Now().Add(240 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "令牌生成失败")
		return
	}

	model.Success(c, "success", gin.H{
		"token": tokenString,
		"user":  user.ToProfile(),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	model.Success(c, "注销成功", nil)
}

type RegisterInput struct {
	Name                 string `json:"name" binding:"required,max=255"`
	Email                string `json:"email" binding:"required,email,max=255"`
	Password             string `json:"password" binding:"required,min=6,max=32"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "请填写有效邮箱，密码长度需为 6-32 位")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	if input.Name == "" {
		model.Fail(c, http.StatusUnprocessableEntity, "用户名不能为空")
		return
	}
	if input.PasswordConfirmation != "" && input.Password != input.PasswordConfirmation {
		model.Fail(c, http.StatusUnprocessableEntity, "两次密码输入不一致")
		return
	}

	// Check if email exists
	var count int64
	config.DB.Model(&model.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "邮箱已被注册")
		return
	}

	passwordHash, _ := config.HashPassword(input.Password)
	initialCapacity := defaultUserCapacity()

	var defaultGroup model.Group
	if err := config.DB.Where("is_default = ?", true).First(&defaultGroup).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "默认角色组未配置")
		return
	}

	user := model.User{
		Name:         input.Name,
		Email:        input.Email,
		Password:     passwordHash,
		RegisteredIP: c.ClientIP(),
		GroupID:      &defaultGroup.ID,
		Configs:      model.JSONMap{"default_permission": 0},
		Capacity:     initialCapacity,
		Status:       1,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "注册失败")
		return
	}

	// Issue token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"is_adminer": user.IsAdminer,
		"exp":        time.Now().Add(240 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	})
	tokenString, _ := token.SignedString([]byte(h.cfg.JWTSecret))

	model.Success(c, "注册成功", gin.H{
		"token": tokenString,
		"user":  user.ToProfile(),
	})
}

func defaultUserCapacity() float64 {
	var cfg model.SystemConfig
	if err := config.DB.Where("name = ?", "user_initial_capacity").First(&cfg).Error; err != nil {
		return 512000
	}
	capacity, err := strconv.ParseFloat(strings.TrimSpace(cfg.Value), 64)
	if err != nil || capacity < 0 {
		return 512000
	}
	return capacity
}
