package handler

import (
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"

	"github.com/gin-gonic/gin"
)

type AlbumHandler struct{}

func NewAlbumHandler() *AlbumHandler { return &AlbumHandler{} }

func (h *AlbumHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	var albums []model.Album
	config.DB.Where("user_id = ?", userID).Order("id DESC").Find(&albums)
	model.Success(c, "success", albums)
}

type AlbumInput struct {
	Name  string `json:"name" binding:"required,max=64"`
	Intro string `json:"intro"`
}

func (h *AlbumHandler) Create(c *gin.Context) {
	var input AlbumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	userID := c.GetUint("user_id")
	album := model.Album{
		UserID:   userID,
		Name:     input.Name,
		Intro:    input.Intro,
		ImageNum: 0,
	}

	if err := config.DB.Create(&album).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}

	// Update user album count
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("album_num", config.DB.Raw("album_num + 1"))

	model.Success(c, "创建成功", album)
}

func (h *AlbumHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var input AlbumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	result := config.DB.Model(&model.Album{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
		"name":  input.Name,
		"intro": input.Intro,
	})
	if result.RowsAffected == 0 {
		model.Fail(c, http.StatusNotFound, "相册不存在")
		return
	}

	model.Success(c, "更新成功", nil)
}

func (h *AlbumHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var album model.Album
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&album).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "相册不存在")
		return
	}

	// Move images in album to no album
	config.DB.Model(&model.Image{}).Where("album_id = ?", album.ID).Update("album_id", nil)

	config.DB.Delete(&album)
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("album_num", config.DB.Raw("CASE WHEN album_num > 0 THEN album_num - 1 ELSE 0 END"))

	model.Success(c, "删除成功", nil)
}
