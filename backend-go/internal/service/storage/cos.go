package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSAdapter 腾讯云对象存储适配器
type COSAdapter struct {
	client    *cos.Client
	bucket    string
	publicURL string
}

// buildCOSBucketHost 根据 bucket、appID、region 构建 COS 访问域名
// 兼容旧格式(带 appID)和新格式(不带 appID)
func buildCOSBucketHost(bucket, appID, region string) string {
	if appID != "" {
		return fmt.Sprintf("%s-%s.cos.%s.myqcloud.com", bucket, appID, region)
	}
	return fmt.Sprintf("%s.cos.%s.myqcloud.com", bucket, region)
}

// NewCOSAdapter 创建 COS 存储适配器
func NewCOSAdapter(secretID, secretKey, region, bucket, appID, publicURL string) *COSAdapter {
	host := buildCOSBucketHost(bucket, appID, region)
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s", host))

	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	if publicURL == "" {
		publicURL = fmt.Sprintf("https://%s", host)
	}

	return &COSAdapter{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

func (a *COSAdapter) Save(path string, data []byte) error {
	_, err := a.client.Object.Put(context.Background(), path, bytes.NewReader(data), nil)
	return err
}

func (a *COSAdapter) Delete(path string) error {
	_, err := a.client.Object.Delete(context.Background(), path)
	return err
}

func (a *COSAdapter) Exists(path string) bool {
	resp, err := a.client.Object.Head(context.Background(), path, nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (a *COSAdapter) URL(path string) string {
	u := a.publicURL
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	return u + path
}

// SaveStream 流式上传（用于大文件）
func (a *COSAdapter) SaveStream(path string, reader io.Reader) error {
	_, err := a.client.Object.Put(context.Background(), path, reader, nil)
	return err
}
