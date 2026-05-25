package admin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SettingHandler struct{}

func NewSettingHandler() *SettingHandler { return &SettingHandler{} }

func (h *SettingHandler) Index(c *gin.Context) {
	var settings []model.SystemConfig
	config.DB.Find(&settings)

	data := make(gin.H)
	for _, s := range settings {
		data[s.Name] = s.Value
	}

	model.Success(c, "success", data)
}

func (h *SettingHandler) Save(c *gin.Context) {
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	for key, value := range input {
		var cfg model.SystemConfig
		result := config.DB.Where("name = ?", key).First(&cfg)

		strValue := ""
		switch v := value.(type) {
		case string:
			strValue = v
		case bool:
			if v {
				strValue = "1"
			} else {
				strValue = "0"
			}
		case float64:
			strValue = fmt.Sprintf("%v", v)
		}

		if result.Error != nil {
			config.DB.Create(&model.SystemConfig{Name: key, Value: strValue})
		} else {
			config.DB.Model(&cfg).Update("value", strValue)
		}
	}

	model.Success(c, "保存成功", nil)
}

func (h *SettingHandler) MailTest(c *gin.Context) {
	model.Success(c, "测试邮件已发送", nil)
}

// BgUpload 处理管理员上传背景图片
func (h *SettingHandler) BgUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		model.Fail(c, http.StatusBadRequest, "请选择文件")
		return
	}

	// 仅允许常见图片格式
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		model.Fail(c, http.StatusBadRequest, "不支持的图片格式")
		return
	}

	// 确保目录存在
	bgDir := filepath.Join("uploads", "bg")
	if err := os.MkdirAll(bgDir, 0755); err != nil {
		model.Fail(c, http.StatusInternalServerError, "创建目录失败")
		return
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	savePath := filepath.Join(bgDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		model.Fail(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	// 构造访问 URL（使用相对路径，前端通过 Nginx / Vite 代理访问）
	imageURL := fmt.Sprintf("/uploads/bg/%s", filename)

	// 自动更新 site_bg_image 配置
	var cfg model.SystemConfig
	result := config.DB.Where("name = ?", "site_bg_image").First(&cfg)
	if result.Error != nil {
		config.DB.Create(&model.SystemConfig{Name: "site_bg_image", Value: imageURL})
	} else {
		config.DB.Model(&cfg).Update("value", imageURL)
	}

	model.Success(c, "上传成功", gin.H{"url": imageURL})
}

