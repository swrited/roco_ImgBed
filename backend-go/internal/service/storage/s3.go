package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Adapter struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewS3Adapter(accessKey, secretKey, region, bucket, endpoint, publicURL string) *S3Adapter {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{URL: endpoint, SigningRegion: region}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)

	if err != nil {
		panic(fmt.Sprintf("S3 客户端创建失败: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true
		}
	})

	if publicURL == "" {
		publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
	}

	return &S3Adapter{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

func (a *S3Adapter) Save(path string, data []byte) error {
	_, err := a.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (a *S3Adapter) Delete(path string) error {
	_, err := a.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (a *S3Adapter) Exists(path string) bool {
	_, err := a.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	return err == nil
}

func (a *S3Adapter) URL(path string) string {
	u := a.publicURL
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	return u + path
}
