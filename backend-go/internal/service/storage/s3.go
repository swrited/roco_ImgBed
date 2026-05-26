package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Adapter struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewS3Adapter(accessKey, secretKey, region, bucket, endpoint, publicURL string) *S3Adapter {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if err != nil {
		panic(fmt.Sprintf("S3 客户端创建失败: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
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

func (a *S3Adapter) Open(path string) (io.ReadCloser, error) {
	output, err := a.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (a *S3Adapter) SetPublic(path string, public bool) error {
	acl := types.ObjectCannedACLPrivate
	if public {
		acl = types.ObjectCannedACLPublicRead
	}
	_, err := a.client.PutObjectAcl(context.Background(), &s3.PutObjectAclInput{
		ACL:    acl,
		Bucket: aws.String(a.bucket),
		Key:    aws.String(path),
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
