package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
	"lskypro-server/internal/service/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageHandler struct{}

func NewImageHandler() *ImageHandler { return &ImageHandler{} }

func generateKey(length int) string {
	if length < 6 {
		length = 6
	}
	return model.RandomString(length)
}

// allowedExts 支持的图片扩展名白名单
var allowedExts = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true,
	"tif": true, "bmp": true, "ico": true, "psd": true, "webp": true,
}

// normalizeExt 统一扩展名格式（如 jpeg -> jpg）
func normalizeExt(ext string) string {
	e := strings.ToLower(strings.TrimPrefix(ext, "."))
	if e == "jpeg" {
		return "jpg"
	}
	return e
}

// extractUserID 从 gin.Context 中提取当前登录用户 ID
func extractUserID(c *gin.Context) *uint {
	uid := c.GetUint("user_id")
	if uid > 0 {
		return &uid
	}
	return nil
}

// readFileData 从上传表单中读取并返回文件内容
func readFileData(file io.Reader) ([]byte, string, string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", "", err
	}
	md5Sum := fmt.Sprintf("%x", md5.Sum(data))
	sha1Sum := fmt.Sprintf("%x", sha1.Sum(data))
	return data, md5Sum, sha1Sum, nil
}

// checkCapacity 检查用户容量是否充足
func checkCapacity(userID uint, sizeKB float64) error {
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil // 未找到用户时不做限制
	}
	if user.Capacity <= 0 {
		return nil // 无限制
	}
	var usedSize float64
	config.DB.Model(&model.Image{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&usedSize)
	if usedSize+sizeKB > user.Capacity {
		return fmt.Errorf("存储空间不足，请升级或清理图片")
	}
	return nil
}

// resolveStrategy 确定当前用户使用的存储策略
func resolveStrategy(userID *uint) (*model.Strategy, error) {
	if userID != nil {
		var user model.User
		if err := config.DB.First(&user, *userID).Error; err != nil {
			return nil, fmt.Errorf("用户不存在")
		}

		// 1. 检查用户偏好 default_strategy
		if ds, ok := user.Configs["default_strategy"]; ok {
			var strategyID uint
			switch v := ds.(type) {
			case float64:
				strategyID = uint(v)
			case int:
				strategyID = uint(v)
			case int64:
				strategyID = uint(v)
			case uint:
				strategyID = v
			}
			if strategyID > 0 {
				var strategy model.Strategy
				if err := config.DB.First(&strategy, strategyID).Error; err == nil {
					return &strategy, nil
				}
			}
		}

		// 2. 使用管理员设置的系统默认策略
		if strategy, ok := resolveSystemDefaultStrategy(); ok {
			return strategy, nil
		}

		// 3. 使用用户组第一个可用策略
		if user.GroupID != nil {
			var gs model.GroupStrategy
			if err := config.DB.Where("group_id = ?", *user.GroupID).First(&gs).Error; err == nil {
				var strategy model.Strategy
				if err := config.DB.First(&strategy, gs.StrategyID).Error; err == nil {
					return &strategy, nil
				}
			}
		}
	}

	if strategy, ok := resolveSystemDefaultStrategy(); ok {
		return strategy, nil
	}

	// 4. 默认本地策略
	var strategy model.Strategy
	if err := config.DB.Where("`key` = ?", model.StrategyLocal).First(&strategy).Error; err != nil {
		return nil, fmt.Errorf("存储策略未配置")
	}
	return &strategy, nil
}

func resolveSystemDefaultStrategy() (*model.Strategy, bool) {
	var cfg model.SystemConfig
	if err := config.DB.Where("name = ?", "default_strategy_id").First(&cfg).Error; err != nil {
		return nil, false
	}
	strategyID, err := strconv.ParseUint(strings.TrimSpace(cfg.Value), 10, 64)
	if err != nil || strategyID == 0 {
		return nil, false
	}
	var strategy model.Strategy
	if err := config.DB.First(&strategy, uint(strategyID)).Error; err != nil {
		return nil, false
	}
	return &strategy, true
}

// buildObjectPath 按 user_id/YYYY/MM/DD/uuid.ext 格式生成存储路径
// 游客上传使用 guest/YYYY/MM/DD/uuid.ext
func buildObjectPath(userID *uint, ext string) (objectPath, dirPath, uniqueName string) {
	now := time.Now()
	uniqueName = strings.ReplaceAll(uuid.New().String(), "-", "") + "." + ext
	prefix := "guest"
	if userID != nil {
		prefix = fmt.Sprintf("user_%d", *userID)
	}
	dirPath = fmt.Sprintf("%s/%04d/%02d/%02d", prefix, now.Year(), now.Month(), now.Day())
	objectPath = dirPath + "/" + uniqueName
	return
}

// buildImageURL 拼接策略基础 URL 与对象路径
func buildImageURL(baseURL, objectPath string) string {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL + objectPath
}

// buildUploadResponse 构建上传成功的响应数据
func buildUploadResponse(img model.Image, url string) gin.H {
	return gin.H{
		"key":         img.Key,
		"name":        img.Name,
		"alias_name":  img.AliasName,
		"pathname":    img.Path + "/" + img.Name,
		"origin_name": img.OriginName,
		"size":        img.Size,
		"mimetype":    img.Mimetype,
		"extension":   img.Extension,
		"md5":         img.MD5,
		"sha1":        img.SHA1,
		"links": model.ImageLinks{
			URL:              url,
			HTML:             fmt.Sprintf(`<img src="%s" />`, url),
			BBCode:           fmt.Sprintf(`[img]%s[/img]`, url),
			Markdown:         fmt.Sprintf(`![](%s)`, url),
			MarkdownWithLink: fmt.Sprintf(`[![](%s)](%s)`, url, url),
			ThumbnailURL:     url,
		},
	}
}

func buildImageLinks(url string) model.ImageLinks {
	return model.ImageLinks{
		URL:              url,
		HTML:             fmt.Sprintf(`<img src="%s" />`, url),
		BBCode:           fmt.Sprintf(`[img]%s[/img]`, url),
		Markdown:         fmt.Sprintf(`![](%s)`, url),
		MarkdownWithLink: fmt.Sprintf(`[![](%s)](%s)`, url, url),
		ThumbnailURL:     url,
	}
}

func buildImageResponse(img model.Image) gin.H {
	url := ""
	if img.StrategyID != nil {
		var strategy model.Strategy
		if err := config.DB.First(&strategy, *img.StrategyID).Error; err == nil {
			url = buildImageURL(storage.GetStrategyURL(&strategy), img.Pathname())
		}
	}
	return gin.H{
		"key":         img.Key,
		"name":        img.Name,
		"alias_name":  img.AliasName,
		"origin_name": img.OriginName,
		"pathname":    img.Pathname(),
		"size":        img.Size,
		"width":       img.Width,
		"height":      img.Height,
		"md5":         img.MD5,
		"sha1":        img.SHA1,
		"links":       buildImageLinks(url),
	}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	// 1. 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "请选择文件")
		return
	}

	// 2. 校验扩展名
	ext := normalizeExt(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		model.Fail(c, http.StatusUnprocessableEntity, "不支持的文件类型")
		return
	}

	sizeKB := float64(file.Size) / 1024.0
	maxUploadSize := systemConfigInt("upload_max_size", 10240)
	if maxUploadSize > 0 && sizeKB > float64(maxUploadSize) {
		model.Fail(c, http.StatusUnprocessableEntity, "图片大小超过限制，最大允许 "+formatKBLimit(maxUploadSize))
		return
	}

	// 3. 读取文件内容
	src, err := file.Open()
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "文件打开失败")
		return
	}
	defer src.Close()

	data, md5Sum, sha1Sum, err := readFileData(src)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "文件读取失败: "+err.Error())
		return
	}

	userID := extractUserID(c)

	// 4. 容量检查
	if userID != nil {
		if err := checkCapacity(*userID, sizeKB); err != nil {
			model.Fail(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
	}

	// 5. 解析存储策略
	strategy, err := resolveStrategy(userID)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 6. 创建存储适配器
	adapter, err := storage.Factory(strategy)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "存储适配器创建失败: "+err.Error())
		return
	}

	// 7. 构建存储路径
	objectPath, dirPath, uniqueName := buildObjectPath(userID, ext)

	// 8. 保存文件到存储后端
	if err := adapter.Save(objectPath, data); err != nil {
		model.Fail(c, http.StatusInternalServerError, "文件保存失败: "+err.Error())
		return
	}

	// 9. 获取策略访问 URL
	strategyURL := storage.GetStrategyURL(strategy)

	// 10. 解析图片宽高
	var imgWidth, imgHeight uint
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		imgWidth = uint(cfg.Width)
		imgHeight = uint(cfg.Height)
	}

	// 11. 写入数据库

	// 11. 确定目标相册与处理 Tags。未指定相册时保持未分类。
	var targetAlbumID *uint
	var tags []model.Tag
	if userID != nil {
		targetAlbumID = resolveRequestedAlbumID(*userID, c.PostForm("album_id"))

		tagNames := c.PostFormArray("tags[]")
		if len(tagNames) == 0 {
			if tagsStr := c.PostForm("tags"); tagsStr != "" {
				tagNames = strings.Split(tagsStr, ",")
			}
		}
		if len(tagNames) > 5 {
			tagNames = tagNames[:5]
		}
		for _, name := range tagNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			var tag model.Tag
			if err := config.DB.Where("name = ? AND user_id = ?", name, *userID).FirstOrCreate(&tag, model.Tag{Name: name, UserID: *userID}).Error; err == nil {
				tags = append(tags, tag)
			}
		}
	}

	image := model.Image{
		UserID:     userID,
		AlbumID:    targetAlbumID,
		StrategyID: &strategy.ID,
		Key:        generateKey(6),
		Path:       dirPath,
		Name:       uniqueName,
		OriginName: file.Filename,
		AliasName:  file.Filename,
		Size:       math.Round(sizeKB*1000) / 1000,
		Mimetype:   file.Header.Get("Content-Type"),
		Extension:  ext,
		MD5:        md5Sum,
		SHA1:       sha1Sum,
		Width:      imgWidth,
		Height:     imgHeight,
		UploadedIP: c.ClientIP(),
		Tags:       tags,
	}

	if err := config.DB.Select("*").Create(&image).Error; err != nil {
		adapter.Delete(objectPath) // 回滚存储
		model.Fail(c, http.StatusInternalServerError, "数据库保存失败")
		return
	}

	// 11. 更新用户图片计数
	if userID != nil {
		config.DB.Model(&model.User{}).Where("id = ?", *userID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))
	}
	if targetAlbumID != nil {
		config.DB.Model(&model.Album{}).Where("id = ?", *targetAlbumID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))
	}

	// 12. 构建响应
	imageURL := buildImageURL(strategyURL, objectPath)
	model.Success(c, "上传成功", buildUploadResponse(image, imageURL))
}

func resolveDefaultAlbumID(userID uint, configs model.JSONMap) *uint {
	if configs == nil {
		return nil
	}
	raw, ok := configs["default_album_id"]
	if !ok {
		return nil
	}
	var albumID uint
	switch v := raw.(type) {
	case float64:
		albumID = uint(v)
	case int:
		albumID = uint(v)
	case int64:
		albumID = uint(v)
	case uint:
		albumID = v
	case string:
		if parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
			albumID = uint(parsed)
		}
	}
	if albumID == 0 {
		return nil
	}
	var album model.Album
	if err := config.DB.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
		return nil
	}
	return &albumID
}

func resolveRequestedAlbumID(userID uint, value string) *uint {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" || value == "null" || value == "__none__" {
		return nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return nil
	}
	albumID := uint(parsed)
	var album model.Album
	if err := config.DB.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
		return nil
	}
	return &albumID
}

func (h *ImageHandler) ListImages(c *gin.Context) {
	userID := c.GetUint("user_id")
	page := 1
	perPage := 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil && pp > 0 && pp <= 100 {
		perPage = pp
	}

	var images []model.Image
	var total int64

	query := config.DB.Model(&model.Image{}).Where("user_id = ?", userID).Preload("Tags")

	if albumID := c.Query("album_id"); albumID != "" {
		if albumID == "0" {
			query = query.Where("album_id IS NULL")
		} else {
			query = query.Where("album_id = ?", albumID)
		}
	}
	if tagID := c.Query("tag_id"); tagID != "" {
		query = query.Joins("JOIN image_tags ON image_tags.image_id = images.id").Where("image_tags.tag_id = ?", tagID)
	}
	if q := c.Query("q"); q != "" {
		query = query.Where("origin_name LIKE ? OR alias_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	sort := c.DefaultQuery("sort", "newest")
	switch sort {
	case "earliest":
		query = query.Order("created_at ASC")
	case "utmost":
		query = query.Order("size DESC")
	case "least":
		query = query.Order("size ASC")
	default:
		query = query.Order("created_at DESC")
	}

	query.Count(&total)
	query.Offset((page - 1) * perPage).Limit(perPage).Find(&images)

	imageDTOs := buildImageDTOs(images)

	model.Success(c, "success", gin.H{
		"data":         imageDTOs,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func formatKBLimit(kb int) string {
	if kb >= 1048576 {
		return fmt.Sprintf("%.2f GB", float64(kb)/1048576)
	}
	if kb >= 1024 {
		return fmt.Sprintf("%.2f MB", float64(kb)/1024)
	}
	return fmt.Sprintf("%d KB", kb)
}

func (h *ImageHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	userID := c.GetUint("user_id")

	var image model.Image
	if err := config.DB.Where("`key` = ? AND user_id = ?", key, userID).First(&image).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "图片不存在")
		return
	}

	// 软删除：只设置 deleted_at，不删除物理文件
	config.DB.Delete(&model.Image{}, "id = ?", image.ID)
	if image.AlbumID != nil {
		config.DB.Model(&model.Album{}).Where("id = ?", *image.AlbumID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))
	}
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))

	model.Success(c, "已移至回收站", nil)
}

func (h *ImageHandler) BatchDelete(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Keys []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.Keys) == 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "请提供要删除的图片")
		return
	}

	var images []model.Image
	config.DB.Where("`key` IN ? AND user_id = ?", input.Keys, userID).Find(&images)
	if len(images) == 0 {
		model.Success(c, "已移至回收站", nil)
		return
	}

	// 收集 ID 列表批量软删除
	ids := make([]uint, 0, len(images))
	for _, img := range images {
		ids = append(ids, img.ID)
	}
	config.DB.Delete(&model.Image{}, "id IN ?", ids)

	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num >= ? THEN image_num - ? ELSE 0 END", len(images), len(images)))

	model.Success(c, "已移至回收站", nil)
}

func (h *ImageHandler) TrashList(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage := 30
	var images []model.Image
	var total int64

	query := config.DB.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", userID)
	query.Model(&model.Image{}).Count(&total)

	if err := query.Order("deleted_at DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&images).Error; err != nil {
		model.Fail(c, http.StatusInternalServerError, "获取回收站列表失败")
		return
	}

	dtos := buildImageDTOs(images)
	model.Success(c, "success", gin.H{
		"data":         dtos,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *ImageHandler) RestoreTrash(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Keys []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.Keys) == 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "请提供要恢复的图片")
		return
	}

	var images []model.Image
	config.DB.Unscoped().Where("`key` IN ? AND user_id = ? AND deleted_at IS NOT NULL", input.Keys, userID).Find(&images)
	if len(images) == 0 {
		model.Success(c, "恢复成功", nil)
		return
	}

	// 恢复记录
	config.DB.Unscoped().Model(&model.Image{}).Where("`key` IN ?", input.Keys).Update("deleted_at", nil)

	// 恢复 user 的 image_num
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("image_num + ?", len(images)))

	// 恢复 album 的 image_num
	albumCounts := make(map[uint]int)
	for _, img := range images {
		if img.AlbumID != nil {
			albumCounts[*img.AlbumID]++
		}
	}
	for aID, count := range albumCounts {
		config.DB.Model(&model.Album{}).Where("id = ?", aID).UpdateColumn("image_num", config.DB.Raw("image_num + ?", count))
	}

	model.Success(c, "恢复成功", nil)
}

func (h *ImageHandler) ForceDeleteTrash(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Keys []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.Keys) == 0 {
		model.Fail(c, http.StatusUnprocessableEntity, "请提供要彻底删除的图片")
		return
	}

	var images []model.Image
	config.DB.Unscoped().Where("`key` IN ? AND user_id = ? AND deleted_at IS NOT NULL", input.Keys, userID).Find(&images)

	for _, img := range images {
		if img.StrategyID != nil {
			var strategy model.Strategy
			if err := config.DB.First(&strategy, *img.StrategyID).Error; err == nil {
				if adapter, err := storage.Factory(&strategy); err == nil {
					adapter.Delete(img.Pathname())
				}
			}
		}
		config.DB.Unscoped().Exec("DELETE FROM image_tags WHERE image_id = ?", img.ID)
		config.DB.Unscoped().Delete(&img)
	}

	model.Success(c, "彻底删除成功", nil)
}

func (h *ImageHandler) Rename(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Key       string `json:"key" binding:"required"`
		AliasName string `json:"alias_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	result := config.DB.Model(&model.Image{}).Where("`key` = ? AND user_id = ?", input.Key, userID).Update("alias_name", input.AliasName)
	if result.RowsAffected == 0 {
		model.Fail(c, http.StatusNotFound, "图片不存在")
		return
	}

	model.Success(c, "重命名成功", nil)
}

func (h *ImageHandler) Move(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Keys    []string `json:"keys" binding:"required"`
		AlbumID *uint    `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	for _, key := range input.Keys {
		var img model.Image
		if err := config.DB.Where("`key` = ? AND user_id = ?", key, userID).First(&img).Error; err != nil {
			continue
		}

		oldAlbumID := img.AlbumID
		config.DB.Model(&img).Update("album_id", input.AlbumID)

		if oldAlbumID != nil {
			config.DB.Model(&model.Album{}).Where("id = ?", *oldAlbumID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))
		}
		if input.AlbumID != nil {
			config.DB.Model(&model.Album{}).Where("id = ?", *input.AlbumID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))
		}
	}

	model.Success(c, "移动成功", nil)
}

func (h *ImageHandler) UpdateTags(c *gin.Context) {
	userID := c.GetUint("user_id")
	key := c.Param("key")

	var input struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	var image model.Image
	if err := config.DB.Where("`key` = ? AND user_id = ?", key, userID).First(&image).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "图片不存在")
		return
	}

	var newTags []model.Tag
	if len(input.Tags) > 5 {
		input.Tags = input.Tags[:5]
	}
	for _, name := range input.Tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var tag model.Tag
		if err := config.DB.Where("name = ? AND user_id = ?", name, userID).FirstOrCreate(&tag, model.Tag{Name: name, UserID: userID}).Error; err == nil {
			newTags = append(newTags, tag)
		}
	}

	config.DB.Model(&image).Association("Tags").Replace(newTags)

	model.Success(c, "更新成功", nil)
}

func (h *ImageHandler) Gallery(c *gin.Context) {
	page := 1
	perPage := 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}

	var albums []model.Album
	var total int64

	query := config.DB.Model(&model.Album{}).Preload("User").Where("permission = ?", 1) // 1 = public

	if q := c.Query("q"); q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}

	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&albums)

	var result []map[string]interface{}
	for _, a := range albums {
		userName := ""
		if a.User != nil {
			userName = a.User.Name
		}

		result = append(result, map[string]interface{}{
			"id":         a.ID,
			"name":       a.Name,
			"intro":      a.Intro,
			"image_num":  a.ImageNum,
			"user_name":  userName,
			"cover_url":  albumCoverURL(a),
			"created_at": a.CreatedAt,
		})
	}

	model.Success(c, "success", gin.H{
		"data":         result,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *ImageHandler) GalleryAlbum(c *gin.Context) {
	albumID := c.Param("id")
	page := 1
	perPage := 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}

	var album model.Album
	if err := config.DB.Where("id = ? AND permission = ?", albumID, 1).First(&album).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "相册不存在或非公开")
		return
	}

	var images []model.Image
	var total int64

	query := config.DB.Model(&model.Image{}).Where("album_id = ?", albumID).Preload("Tags")
	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&images)

	imageDTOs := buildImageDTOs(images)

	model.Success(c, "success", gin.H{
		"album": map[string]interface{}{
			"id":   album.ID,
			"name": album.Name,
		},
		"data":         imageDTOs,
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func (h *ImageHandler) Random(c *gin.Context) {
	query := randomBaseQuery(c)
	query = applyRandomFilters(query, c)

	img, err := pickRandomImage(query)
	if err != nil {
		model.Fail(c, http.StatusNotFound, "没有匹配的图片")
		return
	}

	model.Success(c, "success", buildImageResponse(img))
}

func (h *ImageHandler) Adaptive(c *gin.Context) {
	orientation, ratio := adaptiveTarget(c.GetHeader("User-Agent"))

	img, err := pickRandomImage(applyRatioFilter(applyOrientationFilter(randomBaseQuery(c), orientation), ratio))
	if err != nil {
		img, err = pickRandomImage(applyOrientationFilter(randomBaseQuery(c), orientation))
	}
	if err != nil {
		img, err = pickRandomImage(randomBaseQuery(c))
	}
	if err != nil {
		model.Fail(c, http.StatusNotFound, "没有可返回的图片")
		return
	}

	model.Success(c, "success", buildImageResponse(img))
}

func randomBaseQuery(c *gin.Context) *gorm.DB {
	query := config.DB.Model(&model.Image{}).Where("is_unhealthy = ?", false)
	userID := c.GetUint("user_id")
	if userID > 0 {
		return query.Where("user_id = ?", userID)
	}
	return query.Where("permission = ?", 1)
}

func applyRandomFilters(query *gorm.DB, c *gin.Context) *gorm.DB {
	if albumID := c.Query("album_id"); albumID != "" {
		if albumID == "0" {
			query = query.Where("album_id IS NULL")
		} else {
			query = query.Where("album_id = ?", albumID)
		}
	}
	if orientation := c.Query("orientation"); orientation != "" {
		query = applyOrientationFilter(query, orientation)
	}
	if ratio := c.Query("ratio"); ratio != "" {
		query = applyRatioFilter(query, ratio)
	}
	if v, ok := uintQuery(c, "min_width"); ok {
		query = query.Where("width >= ?", v)
	}
	if v, ok := uintQuery(c, "max_width"); ok {
		query = query.Where("width <= ?", v)
	}
	if v, ok := uintQuery(c, "min_height"); ok {
		query = query.Where("height >= ?", v)
	}
	if v, ok := uintQuery(c, "max_height"); ok {
		query = query.Where("height <= ?", v)
	}
	return query
}

func applyOrientationFilter(query *gorm.DB, orientation string) *gorm.DB {
	switch strings.ToLower(strings.TrimSpace(orientation)) {
	case "landscape":
		return query.Where("width > height")
	case "portrait":
		return query.Where("height > width")
	case "square":
		return query.Where("width = height")
	default:
		return query
	}
}

func applyRatioFilter(query *gorm.DB, ratioText string) *gorm.DB {
	ratio, ok := parseRatio(ratioText)
	if !ok || ratio <= 0 {
		return query
	}
	minRatio := ratio * 0.95
	maxRatio := ratio * 1.05
	return query.Where("height > 0 AND (CAST(width AS REAL) / CAST(height AS REAL)) BETWEEN ? AND ?", minRatio, maxRatio)
}

func parseRatio(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if strings.Contains(value, ":") {
		parts := strings.SplitN(value, ":", 2)
		w, wErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		h, hErr := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if wErr != nil || hErr != nil || h == 0 {
			return 0, false
		}
		return w / h, true
	}
	ratio, err := strconv.ParseFloat(value, 64)
	return ratio, err == nil
}

func uintQuery(c *gin.Context, key string) (uint, bool) {
	value := c.Query(key)
	if value == "" {
		return 0, false
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(n), true
}

func pickRandomImage(query *gorm.DB) (model.Image, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return model.Image{}, err
	}
	if total <= 0 {
		return model.Image{}, errors.New("no image")
	}
	offset := int(time.Now().UnixNano() % total)
	var img model.Image
	if err := query.Offset(offset).Limit(1).Find(&img).Error; err != nil {
		return model.Image{}, err
	}
	if img.ID == 0 {
		return model.Image{}, errors.New("no image")
	}
	return img, nil
}

func adaptiveTarget(userAgent string) (orientation string, ratio string) {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipod"):
		return "portrait", "9:16"
	case strings.Contains(ua, "android") && strings.Contains(ua, "mobile"):
		return "portrait", "9:16"
	case strings.Contains(ua, "ipad"):
		return "landscape", "4:3"
	case strings.Contains(ua, "android"):
		return "landscape", "16:9"
	case strings.Contains(ua, "windows"), strings.Contains(ua, "macintosh"), strings.Contains(ua, "mac os"):
		return "landscape", "16:9"
	default:
		return "landscape", ""
	}
}

// ImageDTO wraps model.Image with a computed URL field
type ImageDTO struct {
	model.Image
	URL   string           `json:"url"`
	Links model.ImageLinks `json:"links"`
}

// buildImageDTOs 为图片列表构建带 URL 的 DTO
func buildImageDTOs(images []model.Image) []ImageDTO {
	if len(images) == 0 {
		return nil
	}

	// 收集唯一的策略 ID
	strategyIDs := make(map[uint]bool)
	for _, img := range images {
		if img.StrategyID != nil {
			strategyIDs[*img.StrategyID] = true
		}
	}

	// 批量加载策略并构建 URL 映射
	strategyURLs := make(map[uint]string)
	if len(strategyIDs) > 0 {
		ids := make([]uint, 0, len(strategyIDs))
		for id := range strategyIDs {
			ids = append(ids, id)
		}

		var strategies []model.Strategy
		config.DB.Where("id IN ?", ids).Find(&strategies)
		for _, s := range strategies {
			strategyURLs[s.ID] = storage.GetStrategyURL(&s)
		}
	}

	// 构建 DTO
	result := make([]ImageDTO, len(images))
	for i, img := range images {
		dto := ImageDTO{Image: img}
		if img.StrategyID != nil {
			if baseURL, ok := strategyURLs[*img.StrategyID]; ok {
				if !strings.HasSuffix(baseURL, "/") {
					baseURL += "/"
				}
				dto.URL = baseURL + img.Pathname()
			}
		}
		dto.Links = buildImageLinks(dto.URL)
		result[i] = dto
	}
	return result
}
