package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdimage "image"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
	"lskypro-server/internal/service/storage"

	"github.com/gin-gonic/gin"
)

type AIImageHandler struct{}

func NewAIImageHandler() *AIImageHandler { return &AIImageHandler{} }

type aiImageRequest struct {
	Prompt          string `json:"prompt"`
	AspectRatio     string `json:"aspect_ratio"`
	Count           int    `json:"count"`
	PromptOptimizer bool   `json:"prompt_optimizer"`
}

type minimaxImageResponse struct {
	ID   string `json:"id"`
	Data struct {
		ImageBase64 []string `json:"image_base64"`
		ImageURLs   []string `json:"image_urls"`
	} `json:"data"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

type openAIImageResponse struct {
	Data []struct {
		Base64 string `json:"b64_json"`
		URL    string `json:"url"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type siliconFlowImageResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	Message string `json:"message"`
}

var allowedAIAspectRatios = map[string]bool{
	"1:1": true, "16:9": true, "4:3": true, "3:2": true,
	"2:3": true, "3:4": true, "9:16": true, "21:9": true,
}

var aiImageRequestTimes sync.Map

func (h *AIImageHandler) Generate(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		model.Fail(c, http.StatusUnauthorized, "请先登录")
		return
	}

	if !systemConfigBool("is_enable_ai_image", false) {
		model.Fail(c, http.StatusForbidden, "AI 生图功能未启用")
		return
	}

	provider := aiImageProvider()
	apiKey := aiImageProviderAPIKey(provider)
	if apiKey == "" {
		model.Fail(c, http.StatusUnprocessableEntity, aiImageProviderName(provider)+" API Key 未配置")
		return
	}

	var input aiImageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		model.Fail(c, http.StatusUnprocessableEntity, "参数错误")
		return
	}
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" {
		model.Fail(c, http.StatusUnprocessableEntity, "请输入提示词")
		return
	}
	if len([]rune(input.Prompt)) > 1500 {
		model.Fail(c, http.StatusUnprocessableEntity, "提示词不能超过 1500 个字符")
		return
	}

	if input.AspectRatio == "" {
		input.AspectRatio = "1:1"
	}
	if !allowedAIAspectRatios[input.AspectRatio] {
		model.Fail(c, http.StatusUnprocessableEntity, "不支持的图片比例")
		return
	}

	maxCount := systemConfigInt("ai_image_max_count", 4)
	if maxCount < 1 || maxCount > 9 {
		maxCount = 4
	}
	if input.Count <= 0 {
		input.Count = 1
	}
	if input.Count > maxCount {
		input.Count = maxCount
	}

	if remaining, ok := checkAIImageRateLimit(userID); !ok {
		model.Fail(c, http.StatusTooManyRequests, fmt.Sprintf("AI 生图请求过于频繁，请 %d 秒后再试", remaining))
		return
	}
	if remaining, ok := checkAIImageDailyLimit(userID); !ok {
		model.Fail(c, http.StatusTooManyRequests, fmt.Sprintf("今日 AI 生图免费额度已用完，每天可生成 %d 次", remaining))
		return
	}

	generated, err := callAIImageProvider(provider, apiKey, input)
	if err != nil {
		model.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	if len(generated) == 0 {
		model.Fail(c, http.StatusBadGateway, aiImageProviderName(provider)+" 未返回图片")
		return
	}

	album, err := getOrCreateAIAlbum(userID)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "AI 相册创建失败")
		return
	}

	strategy, err := resolveStrategy(&userID)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adapter, err := storage.Factory(strategy)
	if err != nil {
		model.Fail(c, http.StatusInternalServerError, "存储适配器创建失败: "+err.Error())
		return
	}
	strategyURL := storage.GetStrategyURL(strategy)

	responses := make([]gin.H, 0, len(generated))
	for index, data := range generated {
		image, url, err := saveGeneratedImage(c, adapter, strategy, strategyURL, album.ID, userID, data, input, index)
		if err != nil {
			model.Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		item := buildUploadResponse(image, url)
		item["album_id"] = album.ID
		item["width"] = image.Width
		item["height"] = image.Height
		responses = append(responses, item)
	}
	config.DB.First(album, album.ID)
	recordAIImageUsage(c, userID, input)

	model.Success(c, "生成成功", gin.H{
		"album":    album,
		"images":   responses,
		"quota":    aiImageQuota(userID),
		"provider": provider,
	})
}

func aiImageProvider() string {
	switch strings.ToLower(strings.TrimSpace(systemConfigString("ai_image_provider", "minimax"))) {
	case "openai", "siliconflow", "compatible":
		return strings.ToLower(strings.TrimSpace(systemConfigString("ai_image_provider", "minimax")))
	default:
		return "minimax"
	}
}

func aiImageProviderName(provider string) string {
	switch provider {
	case "openai":
		return "OpenAI"
	case "siliconflow":
		return "SiliconFlow"
	case "compatible":
		return "OpenAI 兼容接口"
	default:
		return "MiniMax"
	}
}

func aiImageProviderAPIKey(provider string) string {
	switch provider {
	case "openai":
		if key := strings.TrimSpace(systemConfigString("openai_image_api_key", "")); key != "" {
			return key
		}
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	case "siliconflow":
		if key := strings.TrimSpace(systemConfigString("siliconflow_image_api_key", "")); key != "" {
			return key
		}
		return strings.TrimSpace(os.Getenv("SILICONFLOW_API_KEY"))
	case "compatible":
		return strings.TrimSpace(systemConfigString("compatible_image_api_key", ""))
	default:
		if key := strings.TrimSpace(systemConfigString("minimax_api_key", "")); key != "" {
			return key
		}
		return strings.TrimSpace(os.Getenv("MINIMAX_API_KEY"))
	}
}

func callAIImageProvider(provider string, apiKey string, input aiImageRequest) ([][]byte, error) {
	switch provider {
	case "openai":
		return callOpenAIImage(apiKey, input,
			systemConfigString("openai_image_api_endpoint", "https://api.openai.com/v1/images/generations"),
			systemConfigString("openai_image_model", "gpt-image-1.5"),
			"OpenAI",
		)
	case "siliconflow":
		return callSiliconFlowImage(apiKey, input)
	case "compatible":
		return callOpenAIImage(apiKey, input,
			systemConfigString("compatible_image_api_endpoint", ""),
			systemConfigString("compatible_image_model", ""),
			"OpenAI 兼容接口",
		)
	default:
		return callMiniMaxImage(apiKey, input)
	}
}

func checkAIImageRateLimit(userID uint) (int, bool) {
	limitSeconds := systemConfigInt("ai_image_rate_limit_seconds", 30)
	if limitSeconds <= 0 {
		return 0, true
	}

	now := time.Now()
	if lastRaw, ok := aiImageRequestTimes.Load(userID); ok {
		if last, ok := lastRaw.(time.Time); ok {
			remaining := time.Duration(limitSeconds)*time.Second - now.Sub(last)
			if remaining > 0 {
				return int(math.Ceil(remaining.Seconds())), false
			}
		}
	}
	aiImageRequestTimes.Store(userID, now)
	return 0, true
}

func checkAIImageDailyLimit(userID uint) (int, bool) {
	limit := systemConfigInt("ai_image_daily_limit", 10)
	if limit <= 0 {
		return 0, true
	}
	used := aiImageUsageToday(userID)
	return limit, used < int64(limit)
}

func aiImageUsageToday(userID uint) int64 {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int64
	config.DB.Model(&model.AIImageUsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, start).
		Count(&count)
	return count
}

func aiImageQuota(userID uint) gin.H {
	limit := systemConfigInt("ai_image_daily_limit", 10)
	if limit <= 0 {
		return gin.H{"limit": 0, "used": aiImageUsageToday(userID), "remaining": -1}
	}
	used := aiImageUsageToday(userID)
	remaining := int64(limit) - used
	if remaining < 0 {
		remaining = 0
	}
	return gin.H{"limit": limit, "used": used, "remaining": remaining}
}

func recordAIImageUsage(c *gin.Context, userID uint, input aiImageRequest) {
	prompt := input.Prompt
	if len([]rune(prompt)) > 120 {
		prompt = string([]rune(prompt)[:120])
	}
	config.DB.Create(&model.AIImageUsageLog{
		UserID: userID,
		Count:  input.Count,
		Prompt: prompt,
		Ratio:  input.AspectRatio,
		IP:     c.ClientIP(),
	})
}

func callMiniMaxImage(apiKey string, input aiImageRequest) ([][]byte, error) {
	endpoint := strings.TrimSpace(systemConfigString("minimax_api_endpoint", "https://api.minimaxi.com/v1/image_generation"))
	modelName := strings.TrimSpace(systemConfigString("minimax_model", "image-01"))
	if modelName == "" {
		modelName = "image-01"
	}

	payload := map[string]interface{}{
		"model":            modelName,
		"prompt":           input.Prompt,
		"aspect_ratio":     input.AspectRatio,
		"response_format":  "base64",
		"n":                input.Count,
		"prompt_optimizer": input.PromptOptimizer,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MiniMax 请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("MiniMax 响应读取失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("MiniMax 请求失败: HTTP %d %s", resp.StatusCode, truncateString(string(respBody), 240))
	}

	var parsed minimaxImageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("MiniMax 响应解析失败: %w", err)
	}
	if parsed.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("MiniMax 生成失败: %s", parsed.BaseResp.StatusMsg)
	}

	images := make([][]byte, 0, len(parsed.Data.ImageBase64)+len(parsed.Data.ImageURLs))
	for _, raw := range parsed.Data.ImageBase64 {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if idx := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && idx >= 0 {
			raw = raw[idx+1:]
		}
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("MiniMax 图片解码失败: %w", err)
		}
		images = append(images, data)
	}
	for _, imageURL := range parsed.Data.ImageURLs {
		data, err := downloadGeneratedImage(imageURL)
		if err != nil {
			return nil, err
		}
		images = append(images, data)
	}
	return images, nil
}

func callOpenAIImage(apiKey string, input aiImageRequest, endpoint string, modelName string, providerName string) ([][]byte, error) {
	endpoint = strings.TrimSpace(endpoint)
	modelName = strings.TrimSpace(modelName)
	if endpoint == "" || modelName == "" {
		return nil, fmt.Errorf("%s 接口地址或模型未配置", providerName)
	}
	payload := map[string]interface{}{
		"model":  modelName,
		"prompt": input.Prompt,
		"n":      input.Count,
		"size":   openAIImageSize(input.AspectRatio),
	}
	respBody, err := postAIImageRequest(endpoint, apiKey, payload, providerName)
	if err != nil {
		return nil, err
	}
	var parsed openAIImageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("%s 响应解析失败: %w", providerName, err)
	}
	if parsed.Error.Message != "" {
		return nil, fmt.Errorf("%s 生成失败: %s", providerName, parsed.Error.Message)
	}
	images := make([][]byte, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		if strings.TrimSpace(item.Base64) != "" {
			data, err := decodeGeneratedBase64(item.Base64, providerName)
			if err != nil {
				return nil, err
			}
			images = append(images, data)
			continue
		}
		if strings.TrimSpace(item.URL) != "" {
			data, err := downloadGeneratedImage(item.URL)
			if err != nil {
				return nil, err
			}
			images = append(images, data)
		}
	}
	return images, nil
}

func callSiliconFlowImage(apiKey string, input aiImageRequest) ([][]byte, error) {
	endpoint := strings.TrimSpace(systemConfigString("siliconflow_image_api_endpoint", "https://api.siliconflow.cn/v1/images/generations"))
	modelName := strings.TrimSpace(systemConfigString("siliconflow_image_model", "Kwai-Kolors/Kolors"))
	if endpoint == "" || modelName == "" {
		return nil, fmt.Errorf("SiliconFlow 接口地址或模型未配置")
	}
	payload := map[string]interface{}{
		"model":      modelName,
		"prompt":     input.Prompt,
		"image_size": siliconFlowImageSize(input.AspectRatio),
		"batch_size": input.Count,
	}
	respBody, err := postAIImageRequest(endpoint, apiKey, payload, "SiliconFlow")
	if err != nil {
		return nil, err
	}
	var parsed siliconFlowImageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("SiliconFlow 响应解析失败: %w", err)
	}
	if len(parsed.Images) == 0 && parsed.Message != "" {
		return nil, fmt.Errorf("SiliconFlow 生成失败: %s", parsed.Message)
	}
	images := make([][]byte, 0, len(parsed.Images))
	for _, item := range parsed.Images {
		if strings.TrimSpace(item.URL) == "" {
			continue
		}
		data, err := downloadGeneratedImage(item.URL)
		if err != nil {
			return nil, err
		}
		images = append(images, data)
	}
	return images, nil
}

func postAIImageRequest(endpoint string, apiKey string, payload map[string]interface{}, providerName string) ([]byte, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s 请求失败: %w", providerName, err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s 响应读取失败: %w", providerName, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s 请求失败: HTTP %d %s", providerName, resp.StatusCode, truncateString(string(respBody), 240))
	}
	return respBody, nil
}

func decodeGeneratedBase64(raw string, providerName string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && idx >= 0 {
		raw = raw[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%s 图片解码失败: %w", providerName, err)
	}
	return data, nil
}

func openAIImageSize(ratio string) string {
	switch ratio {
	case "16:9", "4:3", "3:2", "21:9":
		return "1536x1024"
	case "2:3", "3:4", "9:16":
		return "1024x1536"
	default:
		return "1024x1024"
	}
}

func siliconFlowImageSize(ratio string) string {
	switch ratio {
	case "16:9":
		return "1280x720"
	case "4:3":
		return "1152x864"
	case "3:2":
		return "1248x832"
	case "2:3":
		return "832x1248"
	case "3:4":
		return "864x1152"
	case "9:16":
		return "720x1280"
	case "21:9":
		return "1344x576"
	default:
		return "1024x1024"
	}
}

func downloadGeneratedImage(imageURL string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("图片下载失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("图片下载失败: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 25*1024*1024))
}

func saveGeneratedImage(c *gin.Context, adapter storage.Adapter, strategy *model.Strategy, strategyURL string, albumID uint, userID uint, data []byte, input aiImageRequest, index int) (model.Image, string, error) {
	if len(data) == 0 {
		return model.Image{}, "", fmt.Errorf("生成图片为空")
	}

	sizeKB := float64(len(data)) / 1024.0
	if err := checkCapacity(userID, sizeKB); err != nil {
		return model.Image{}, "", err
	}

	mimeType := http.DetectContentType(data)
	ext := extensionFromMime(mimeType)
	objectPath, dirPath, uniqueName := buildObjectPath(&userID, ext)
	if err := adapter.Save(objectPath, data); err != nil {
		return model.Image{}, "", fmt.Errorf("生成图片保存失败: %w", err)
	}

	md5Sum, sha1Sum := computeHashes(data)
	var width, height uint
	if cfg, _, err := stdimage.DecodeConfig(bytes.NewReader(data)); err == nil {
		width = uint(cfg.Width)
		height = uint(cfg.Height)
	}

	originName := fmt.Sprintf("ai-%s-%d.%s", time.Now().Format("20060102150405"), index+1, ext)
	image := model.Image{
		UserID:     &userID,
		AlbumID:    &albumID,
		StrategyID: &strategy.ID,
		Key:        generateKey(6),
		Path:       dirPath,
		Name:       uniqueName,
		OriginName: originName,
		AliasName:  originName,
		Size:       math.Round(sizeKB*1000) / 1000,
		Mimetype:   mimeType,
		Extension:  ext,
		MD5:        md5Sum,
		SHA1:       sha1Sum,
		Width:      width,
		Height:     height,
		UploadedIP: c.ClientIP(),
	}

	if err := config.DB.Select("*").Create(&image).Error; err != nil {
		_ = adapter.Delete(objectPath)
		return model.Image{}, "", fmt.Errorf("数据库保存失败")
	}
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))
	config.DB.Model(&model.Album{}).Where("id = ?", albumID).UpdateColumn("image_num", config.DB.Raw("image_num + 1"))

	return image, buildImageURL(strategyURL, objectPath), nil
}

func getOrCreateAIAlbum(userID uint) (*model.Album, error) {
	var album model.Album
	err := config.DB.Where("user_id = ? AND name = ?", userID, "AI 生成").First(&album).Error
	if err == nil {
		return &album, nil
	}
	album = model.Album{
		UserID: userID,
		Name:   "AI 生成",
		Intro:  "AI 自动生成的图片",
	}
	if err := config.DB.Create(&album).Error; err != nil {
		return nil, err
	}
	config.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("album_num", config.DB.Raw("album_num + 1"))
	return &album, nil
}

func extensionFromMime(mimeType string) string {
	switch strings.ToLower(strings.Split(mimeType, ";")[0]) {
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	default:
		return "jpg"
	}
}

func computeHashes(data []byte) (string, string) {
	md5Sum := fmt.Sprintf("%x", md5.Sum(data))
	sha1Sum := fmt.Sprintf("%x", sha1.Sum(data))
	return md5Sum, sha1Sum
}

func systemConfigString(key string, fallback string) string {
	var cfg model.SystemConfig
	if err := config.DB.Where("name = ?", key).First(&cfg).Error; err != nil {
		return fallback
	}
	return cfg.Value
}

func systemConfigBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(systemConfigString(key, "")))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes"
}

func systemConfigInt(key string, fallback int) int {
	value := strings.TrimSpace(systemConfigString(key, ""))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func truncateString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
