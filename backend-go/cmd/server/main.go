package main

import (
	"fmt"
	"log"

	"lskypro-server/internal/config"
	"lskypro-server/internal/model"
	"lskypro-server/internal/router"

	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	config.InitDB(cfg)
	autoMigrate(config.DB)

	r := router.Setup(cfg)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("服务器启动中: http://localhost%s", addr)
	r.Run(addr)
}

func autoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Group{},
		&model.User{},
		&model.PasswordReset{},
		&model.Strategy{},
		&model.GroupStrategy{},
		&model.Album{},
		&model.Image{},
		&model.SystemConfig{},
		&model.ApiKey{},
		&model.ApiUsageLog{},
		&model.AIImageUsageLog{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
	seedIfEmpty(db)
	ensureDefaultConfigs(db)
}

func seedIfEmpty(db *gorm.DB) {
	var count int64
	db.Model(&model.Group{}).Count(&count)
	if count == 0 {
		log.Println("初始化默认数据...")

		// Create default group
		group := model.Group{
			Name:      "默认用户组",
			IsDefault: true,
			Configs:   model.JSONMap{},
		}
		db.Create(&group)

		// Create default strategy
		strategy := model.Strategy{
			Key:     model.StrategyLocal,
			Name:    "默认本地策略",
			Intro:   "系统默认的本地策略",
			Configs: model.JSONMap{"url": "http://localhost:8000/i"},
		}
		db.Create(&strategy)

		// Link group and strategy
		db.Create(&model.GroupStrategy{
			GroupID:    group.ID,
			StrategyID: strategy.ID,
		})

		// Create default admin user (admin@admin.com / 123456)
		passwordHash, _ := config.HashPassword("123456")
		adminUser := model.User{
			Name:      "管理员",
			Email:     "admin@admin.com",
			Password:  passwordHash,
			IsAdminer: true,
			GroupID:   &group.ID,
			Capacity:  0, // unlimited
		}
		db.Create(&adminUser)
		log.Println("默认管理员已创建: admin@admin.com / 123456")

		for _, c := range defaultSystemConfigs() {
			db.Create(&c)
		}
	}
}

func ensureDefaultConfigs(db *gorm.DB) {
	for _, item := range defaultSystemConfigs() {
		var cfg model.SystemConfig
		if err := db.Where("name = ?", item.Name).First(&cfg).Error; err != nil {
			db.Create(&item)
		}
	}
}

func defaultSystemConfigs() []model.SystemConfig {
	return []model.SystemConfig{
		{Name: "app_name", Value: "星诺图床"},
		{Name: "app_version", Value: "V 2.1"},
		{Name: "site_description", Value: ""},
		{Name: "site_keywords", Value: ""},
		{Name: "is_enable_registration", Value: "1"},
		{Name: "is_enable_api", Value: "1"},
		{Name: "is_enable_gallery", Value: "1"},
		{Name: "is_allow_guest_upload", Value: "0"},
		{Name: "is_user_need_verify", Value: "0"},
		{Name: "user_initial_capacity", Value: "512000"},
		{Name: "upload_max_size", Value: "10240"},
		{Name: "default_strategy_id", Value: ""},
		{Name: "is_enable_ai_image", Value: "0"},
		{Name: "minimax_api_key", Value: ""},
		{Name: "minimax_api_endpoint", Value: "https://api.minimaxi.com/v1/image_generation"},
		{Name: "minimax_model", Value: "image-01"},
		{Name: "ai_image_max_count", Value: "4"},
		{Name: "ai_image_rate_limit_seconds", Value: "30"},
		{Name: "ai_image_daily_limit", Value: "10"},
		{Name: "api_key_minute_limit", Value: "60"},
		{Name: "api_key_daily_limit", Value: "1000"},
		{Name: "mail", Value: "{}"},
		{Name: "site_bg_image", Value: ""},
		{Name: "site_bg_opacity", Value: "85"},
	}
}
