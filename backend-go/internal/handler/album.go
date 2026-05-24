package handler

import (
	"net/http"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
	"lskypro-server/internal/service/storage"

	"github.com/gin-gonic/gin"
)

type AlbumHandler struct{}

func NewAlbumHandler() *AlbumHandler { return &AlbumHandler{} }

func (h *AlbumHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	var albums []model.Album
	config.DB.Where("user_id = ?", userID).Order("id DESC").Find(&albums)

	data := make([]gin.H, 0, len(albums))
	for _, album := range albums {
		data = append(data, gin.H{
			"id":             album.ID,
			"name":           album.Name,
			"intro":          album.Intro,
			"image_num":      album.ImageNum,
			"permission":     album.Permission,
			"cover_image_id": album.CoverImageID,
			"cover_url":      albumCoverURL(album),
			"created_at":     album.CreatedAt,
			"updated_at":     album.UpdatedAt,
		})
	}
	model.Success(c, "success", data)
}

func albumCoverURL(album model.Album) string {
	var img model.Image
	if album.CoverImageID != nil {
		if err := config.DB.Where("id = ? AND album_id = ?", *album.CoverImageID, album.ID).First(&img).Error; err == nil {
			return imageURL(img)
		}
	}
	if err := config.DB.Where("album_id = ?", album.ID).Order("created_at DESC").First(&img).Error; err != nil {
		return ""
	}
	return imageURL(img)
}

func imageURL(img model.Image) string {
	if img.StrategyID == nil {
		return ""
	}
	var strategy model.Strategy
	if err := config.DB.First(&strategy, *img.StrategyID).Error; err != nil {
		return ""
	}
	return buildImageURL(storage.GetStrategyURL(&strategy), img.Pathname())
}

func validAlbumCoverID(userID uint, albumID uint, coverImageID *uint) *uint {
	if coverImageID == nil || *coverImageID == 0 || albumID == 0 {
		return nil
	}
	var img model.Image
	if err := config.DB.Where("id = ? AND user_id = ? AND album_id = ?", *coverImageID, userID, albumID).First(&img).Error; err != nil {
		return nil
	}
	return coverImageID
}

type AlbumInput struct {
	Name         string `json:"name" binding:"required,max=64"`
	Intro        string `json:"intro"`
	Permission   uint   `json:"permission"`
	CoverImageID *uint  `json:"cover_image_id"`
}

func (h *AlbumHandler) Create(c *gin.Context) {
	var input AlbumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	userID := c.GetUint("user_id")
	album := model.Album{
		UserID:       userID,
		Name:         input.Name,
		Intro:        input.Intro,
		Permission:   input.Permission,
		CoverImageID: validAlbumCoverID(userID, 0, input.CoverImageID),
		ImageNum:     0,
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

	var album model.Album
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&album).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "相册不存在")
		return
	}

	result := config.DB.Model(&model.Album{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
		"name":           input.Name,
		"intro":          input.Intro,
		"permission":     input.Permission,
		"cover_image_id": validAlbumCoverID(userID, album.ID, input.CoverImageID),
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
