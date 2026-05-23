package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// -------- Response DTOs --------

type APIResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PaginatedData struct {
	CurrentPage  int64       `json:"current_page"`
	Data         interface{} `json:"data"`
	FirstPageURL string      `json:"first_page_url"`
	From         int64       `json:"from"`
	LastPage     int64       `json:"last_page"`
	LastPageURL  string      `json:"last_page_url"`
	NextPageURL  string      `json:"next_page_url"`
	Path         string      `json:"path"`
	PerPage      int64       `json:"per_page"`
	PrevPageURL  string      `json:"prev_page_url"`
	To           int64       `json:"to"`
	Total        int64       `json:"total"`
}

// JSON is a custom type for JSON columns
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			bytes = []byte(s)
		} else {
			return errors.New("failed to scan JSONMap: unsupported type")
		}
	}
	return json.Unmarshal(bytes, j)
}

// JSONArr is a custom type for JSON array columns
type JSONArr []interface{}

func (j JSONArr) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONArr) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArr, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			bytes = []byte(s)
		} else {
			return errors.New("failed to scan JSONArr: unsupported type")
		}
	}
	return json.Unmarshal(bytes, j)
}

// -------- Base Model --------

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// -------- User (users) --------

type User struct {
	BaseModel
	GroupID         *uint      `gorm:"column:group_id;default:null" json:"group_id"`
	Name            string     `gorm:"column:name;type:varchar(255)" json:"name"`
	Email           string     `gorm:"column:email;type:varchar(255);uniqueIndex" json:"email"`
	Password        string     `gorm:"column:password;type:varchar(255)" json:"-"`
	RememberToken   string     `gorm:"column:remember_token;type:varchar(100)" json:"-"`
	IsAdminer       bool       `gorm:"column:is_adminer;default:false" json:"is_adminer"`
	Capacity        float64    `gorm:"column:capacity;type:decimal(20,2);default:0" json:"capacity"`
	URL             string     `gorm:"column:url;type:varchar(255);default:''" json:"url"`
	Configs         JSONMap    `gorm:"column:configs;type:json;default:'{}'" json:"-"`
	ImageNum        uint64     `gorm:"column:image_num;default:0" json:"image_num"`
	AlbumNum        uint64     `gorm:"column:album_num;default:0" json:"album_num"`
	RegisteredIP    string     `gorm:"column:registered_ip;type:varchar(45);default:''" json:"registered_ip"`
	Status          uint       `gorm:"column:status;default:1" json:"status"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"email_verified_at"`
	Group           *Group     `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Albums          []Album    `gorm:"foreignKey:UserID" json:"-"`
	Images          []Image    `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string { return "users" }

type UserProfile struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	IsAdminer bool    `json:"is_adminer"`
	Capacity  float64 `json:"capacity"`
	ImageNum  uint64  `json:"image_num"`
	AlbumNum  uint64  `json:"album_num"`
	URL       string  `json:"url"`
	Avatar    string  `json:"avatar"`
}

func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		IsAdminer: u.IsAdminer,
		Capacity:  u.Capacity,
		ImageNum:  u.ImageNum,
		AlbumNum:  u.AlbumNum,
		URL:       u.URL,
		Avatar:    fmt.Sprintf("https://cravatar.cn/avatar/%x?s=96&d=mp&r=g", md5Hash(u.Email)),
	}
}

// -------- Group (groups) --------

type Group struct {
	BaseModel
	Name       string     `gorm:"column:name;type:varchar(64)" json:"name"`
	IsDefault  bool       `gorm:"column:is_default;default:false" json:"is_default"`
	IsGuest    bool       `gorm:"column:is_guest;default:false" json:"is_guest"`
	Configs    JSONMap    `gorm:"column:configs;type:json;default:'{}'" json:"configs"`
	Strategies []Strategy `gorm:"many2many:group_strategy;" json:"strategies,omitempty"`
}

func (Group) TableName() string { return "groups" }

// -------- Strategy (strategies) --------

type Strategy struct {
	BaseModel
	Key     uint    `gorm:"column:key;type:tinyint unsigned" json:"key"`
	Name    string  `gorm:"column:name;type:varchar(64)" json:"name"`
	Intro   string  `gorm:"column:intro;type:varchar(255);default:''" json:"intro"`
	Configs JSONMap `gorm:"column:configs;type:json;default:'{}'" json:"configs"`
}

func (Strategy) TableName() string { return "strategies" }

// -------- Strategy Key Constants --------

const (
	StrategyLocal  uint = 1
	StrategyS3     uint = 2
	StrategyOss    uint = 3
	StrategyCos    uint = 4
	StrategyKodo   uint = 5
	StrategyUss    uint = 6
	StrategySftp   uint = 7
	StrategyFtp    uint = 8
	StrategyWebDav uint = 9
	StrategyMinio  uint = 10
)

var StrategyKeyNames = map[uint]string{
	1: "本地", 2: "AWS S3", 3: "阿里云 OSS", 4: "腾讯云 COS",
	5: "七牛云 Kodo", 6: "又拍云 USS", 7: "SFTP", 8: "FTP",
	9: "WebDav", 10: "Minio",
}

// -------- Album (albums) --------

type Album struct {
	BaseModel
	UserID   uint   `gorm:"column:user_id;index" json:"user_id"`
	Name     string `gorm:"column:name;type:varchar(64)" json:"name"`
	Intro    string `gorm:"column:intro;type:varchar(255);default:''" json:"intro"`
	ImageNum uint64 `gorm:"column:image_num;default:0" json:"image_num"`
	User     *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Album) TableName() string { return "albums" }

// -------- Image (images) --------

type Image struct {
	BaseModel
	UserID      *uint   `gorm:"column:user_id;index;default:null" json:"user_id"`
	AlbumID     *uint   `gorm:"column:album_id;index;default:null" json:"album_id"`
	GroupID     *uint   `gorm:"column:group_id;default:null" json:"group_id"`
	StrategyID  *uint   `gorm:"column:strategy_id;default:null" json:"strategy_id"`
	Key         string  `gorm:"column:key;type:varchar(64);uniqueIndex" json:"key"`
	Path        string  `gorm:"column:path;type:varchar(255)" json:"path"`
	Name        string  `gorm:"column:name;type:varchar(255)" json:"name"`
	OriginName  string  `gorm:"column:origin_name;type:varchar(255)" json:"origin_name"`
	AliasName   string  `gorm:"column:alias_name;type:varchar(255);default:''" json:"alias_name"`
	Size        float64 `gorm:"column:size;type:decimal(10,2);default:0" json:"size"`
	Mimetype    string  `gorm:"column:mimetype;type:varchar(32)" json:"mimetype"`
	Extension   string  `gorm:"column:extension;type:varchar(32)" json:"extension"`
	MD5         string  `gorm:"column:md5;type:varchar(32)" json:"md5"`
	SHA1        string  `gorm:"column:sha1;type:varchar(128)" json:"sha1"`
	Width       uint    `gorm:"column:width;default:0" json:"width"`
	Height      uint    `gorm:"column:height;default:0" json:"height"`
	Permission  uint    `gorm:"column:permission;default:0" json:"permission"`
	IsUnhealthy bool    `gorm:"column:is_unhealthy;default:false" json:"is_unhealthy"`
	UploadedIP  string  `gorm:"column:uploaded_ip;type:varchar(45);default:''" json:"uploaded_ip"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Album    *Album    `gorm:"foreignKey:AlbumID" json:"album,omitempty"`
	Strategy *Strategy `gorm:"foreignKey:StrategyID" json:"strategy,omitempty"`
}

func (Image) TableName() string { return "images" }

func (img *Image) Filename() string {
	if img.AliasName != "" {
		return img.AliasName
	}
	return img.OriginName
}

func (img *Image) Pathname() string {
	return img.Path + "/" + img.Name
}

type ImageLinks struct {
	URL              string `json:"url"`
	HTML             string `json:"html"`
	BBCode           string `json:"bbcode"`
	Markdown         string `json:"markdown"`
	MarkdownWithLink string `json:"markdown_with_link"`
	ThumbnailURL     string `json:"thumbnail_url"`
}

// -------- Config (configs table, key-value) --------

type SystemConfig struct {
	Name  string `gorm:"column:name;type:varchar(32);primaryKey" json:"name"`
	Value string `gorm:"column:value;type:longtext" json:"value"`
}

func (SystemConfig) TableName() string { return "configs" }

// -------- Password Reset --------

type PasswordReset struct {
	Email     string    `gorm:"column:email;index" json:"email"`
	Token     string    `gorm:"column:token;type:varchar(255)" json:"token"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PasswordReset) TableName() string { return "password_resets" }

// -------- ApiKey (api_keys) --------

type ApiKey struct {
	BaseModel
	UserID    uint       `gorm:"column:user_id;index" json:"user_id"`
	Name      string     `gorm:"column:name;type:varchar(64)" json:"name"`
	Key       string     `gorm:"column:key;type:varchar(64);uniqueIndex" json:"key"`
	LastUsed  *time.Time `gorm:"column:last_used" json:"last_used"`
	RevokedAt *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"-"`
}

func (ApiKey) TableName() string { return "api_keys" }

// -------- API Usage Logs --------

type ApiUsageLog struct {
	BaseModel
	UserID    *uint   `gorm:"column:user_id;index" json:"user_id"`
	ApiKeyID  *uint   `gorm:"column:api_key_id;index" json:"api_key_id"`
	Method    string  `gorm:"column:method;type:varchar(16);index" json:"method"`
	Path      string  `gorm:"column:path;type:varchar(255);index" json:"path"`
	Status    int     `gorm:"column:status;index" json:"status"`
	LatencyMS int64   `gorm:"column:latency_ms" json:"latency_ms"`
	IP        string  `gorm:"column:ip;type:varchar(45)" json:"ip"`
	UserAgent string  `gorm:"column:user_agent;type:varchar(255)" json:"user_agent"`
	AuthType  string  `gorm:"column:auth_type;type:varchar(16);index" json:"auth_type"`
	User      *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ApiKey    *ApiKey `gorm:"foreignKey:ApiKeyID" json:"api_key,omitempty"`
}

func (ApiUsageLog) TableName() string { return "api_usage_logs" }

// -------- AI Image Usage Logs --------

type AIImageUsageLog struct {
	BaseModel
	UserID uint   `gorm:"column:user_id;index" json:"user_id"`
	Count  int    `gorm:"column:count;default:1" json:"count"`
	Prompt string `gorm:"column:prompt;type:varchar(255)" json:"prompt"`
	Ratio  string `gorm:"column:ratio;type:varchar(16)" json:"ratio"`
	IP     string `gorm:"column:ip;type:varchar(45)" json:"ip"`
	User   *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AIImageUsageLog) TableName() string { return "ai_image_usage_logs" }

// -------- Group-Strategy Pivot --------

type GroupStrategy struct {
	GroupID    uint `gorm:"column:group_id;primaryKey"`
	StrategyID uint `gorm:"column:strategy_id;primaryKey"`
}

func (GroupStrategy) TableName() string { return "group_strategy" }
