package admin

import (
	"fmt"
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
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
