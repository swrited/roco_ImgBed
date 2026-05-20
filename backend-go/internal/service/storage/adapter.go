package storage

import (
	"fmt"
	"strings"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
)

// Adapter 存储适配器接口
type Adapter interface {
	Save(path string, data []byte) error
	Delete(path string) error
	Exists(path string) bool
	URL(path string) string
}

// requiredMissing 收集配置中缺失的必填字段
func requiredMissing(cfg model.JSONMap, keys ...string) []string {
	var missing []string
	for _, k := range keys {
		if getString(cfg, k, "") == "" {
			missing = append(missing, k)
		}
	}
	return missing
}

// Factory 根据策略创建对应的存储适配器
func Factory(strategy *model.Strategy) (Adapter, error) {
	cfg := strategy.Configs

	switch strategy.Key {
	case model.StrategyLocal:
		root := getString(cfg, "root", "uploads")
		url := getString(cfg, "url", "")
		return NewLocalAdapter(root, url), nil

	case model.StrategyCos:
		if missing := requiredMissing(cfg, "secret_id", "secret_key", "bucket"); len(missing) > 0 {
			return nil, fmt.Errorf("COS 配置不完整，缺少必填字段: %s", strings.Join(missing, ", "))
		}
		return NewCOSAdapter(
			getString(cfg, "secret_id", ""),
			getString(cfg, "secret_key", ""),
			getString(cfg, "region", "ap-guangzhou"),
			getString(cfg, "bucket", ""),
			getString(cfg, "app_id", ""),
			getString(cfg, "url", ""),
		), nil

	case model.StrategyS3:
		if missing := requiredMissing(cfg, "access_key_id", "secret_access_key", "bucket"); len(missing) > 0 {
			return nil, fmt.Errorf("S3 配置不完整，缺少必填字段: %s", strings.Join(missing, ", "))
		}
		return NewS3Adapter(
			getString(cfg, "access_key_id", ""),
			getString(cfg, "secret_access_key", ""),
			getString(cfg, "region", "us-east-1"),
			getString(cfg, "bucket", ""),
			getString(cfg, "endpoint", ""),
			getString(cfg, "url", ""),
		), nil

	default:
		return nil, fmt.Errorf("不支持的存储策略类型: %d (%s)", strategy.Key, model.StrategyKeyNames[strategy.Key])
	}
}

// GetStrategyURL 返回策略的访问 URL
func GetStrategyURL(strategy *model.Strategy) string {
	url := getString(strategy.Configs, "url", config.Get().AppURL)
	queries := getString(strategy.Configs, "queries", "")
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	if queries != "" {
		if !strings.Contains(url, "?") {
			url += "?" + queries
		} else {
			url += "&" + queries
		}
	}
	return url
}

func getString(m model.JSONMap, key, fallback string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return fallback
}
