package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
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

		// 2. 使用用户组第一个可用策略
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

	// 3. 默认本地策略
	var strategy model.Strategy
	if err := config.DB.Where("`key` = ?", model.StrategyLocal).First(&strategy).Error; err != nil {
		return nil, fmt.Errorf("存储策略未配置")
	}
	return &strategy, nil
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
	image := model.Image{
		UserID:     userID,
		StrategyID: &strategy.ID,
		Key:        generateKey(6),
		Path:       dirPath,
		Name:       uniqueName,
		OriginName: file.Filename,
		Size:       math.Round(sizeKB*1000) / 1000,
		Mimetype:   file.Header.Get("Content-Type"),
		Extension:  ext,
		MD5:        md5Sum,
		SHA1:       sha1Sum,
		Width:      imgWidth,
		Height:     imgHeight,
		Permission: 1,
		UploadedIP: c.ClientIP(),
	}

	if err := config.DB.Create(&image).Error; err != nil {
		adapter.Delete(objectPath) // 回滚存储
		model.Fail(c, http.StatusInternalServerError, "数据库保存失败")
		return
	}

	// 11. 更新用户图片计数
	if userID != nil {
		config.DB.Model(&model.User{}).Where("id = ?", *userID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))
	}

	// 12. 构建响应
	imageURL := buildImageURL(strategyURL, objectPath)
	model.Success(c, "上传成功", buildUploadResponse(image, imageURL))
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

	query := config.DB.Model(&model.Image{}).Where("user_id = ?", userID)

	if albumID := c.Query("album_id"); albumID != "" {
		query = query.Where("album_id = ?", albumID)
	}
	if permission := c.Query("permission"); permission != "" {
		query = query.Where("permission = ?", permission)
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

func (h *ImageHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	userID := c.GetUint("user_id")

	var image model.Image
	if err := config.DB.Where("`key` = ? AND user_id = ?", key, userID).First(&image).Error; err != nil {
		model.Fail(c, http.StatusNotFound, "图片不存在")
		return
	}

	// 通过存储适配器删除物理文件
	if image.StrategyID != nil {
		var strategy model.Strategy
		if err := config.DB.First(&strategy, *image.StrategyID).Error; err == nil {
			if adapter, err := storage.Factory(&strategy); err == nil {
				adapter.Delete(image.Pathname())
			}
		}
	}

	config.DB.Delete(&image)
	if image.AlbumID != nil {
		config.DB.Model(&model.Album{}).Where("id = ?", *image.AlbumID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))
	}
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num > 0 THEN image_num - 1 ELSE 0 END"))

	model.Success(c, "删除成功", nil)
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
	for _, img := range images {
		// 通过存储适配器删除物理文件
		if img.StrategyID != nil {
			var strategy model.Strategy
			if err := config.DB.First(&strategy, *img.StrategyID).Error; err == nil {
				if adapter, err := storage.Factory(&strategy); err == nil {
					adapter.Delete(img.Pathname())
				}
			}
		}
		config.DB.Delete(&img)
	}
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("CASE WHEN image_num >= ? THEN image_num - ? ELSE 0 END", len(images), len(images)))

	model.Success(c, "删除成功", nil)
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

func (h *ImageHandler) Permission(c *gin.Context) {
	userID := c.GetUint("user_id")
	var input struct {
		Keys       []string `json:"keys" binding:"required"`
		Permission uint     `json:"permission"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}

	config.DB.Model(&model.Image{}).Where("`key` IN ? AND user_id = ?", input.Keys, userID).Update("permission", input.Permission)

	model.Success(c, "设置成功", nil)
}

func (h *ImageHandler) Gallery(c *gin.Context) {
	page := 1
	perPage := 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}

	var images []model.Image
	var total int64

	query := config.DB.Model(&model.Image{}).Where("permission = ? AND is_unhealthy = ?", 1, false)

	if q := c.Query("q"); q != "" {
		query = query.Where("origin_name LIKE ?", "%"+q+"%")
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query.Order("created_at DESC").Count(&total)
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

// ImageDTO wraps model.Image with a computed URL field
type ImageDTO struct {
	model.Image
	URL string `json:"url"`
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
		result[i] = dto
	}
	return result
}
