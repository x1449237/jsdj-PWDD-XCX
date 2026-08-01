package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// OSSSigner OSS 临时签名 URL 生成器
// 实现阿里云 OSS 签名版本 v1(HMAC-SHA1)，用于生成临时下载/上传 URL
type OSSSigner struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string // 例: oss-cn-hangzhou.aliyuncs.com
	Bucket     string
	CDNDomain  string // CDN 域名(为空则使用 bucket.endpoint)
	SignExpire int    // 签名有效期(秒)
}

// NewOSSSigner 创建 OSS 签名器
func NewOSSSigner(accessKey, secretKey, endpoint, bucket, cdnDomain string, signExpire int) *OSSSigner {
	if signExpire <= 0 {
		signExpire = 300 // 默认 300 秒
	}
	return &OSSSigner{
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		Endpoint:   endpoint,
		Bucket:     bucket,
		CDNDomain:  cdnDomain,
		SignExpire: signExpire,
	}
}

// SignDownloadURL 生成临时下载 URL(默认 300 秒有效期)
// objectKey OSS 对象 key
func (s *OSSSigner) SignDownloadURL(objectKey string) (string, error) {
	return s.SignDownloadURLExpire(objectKey, s.SignExpire)
}

// SignDownloadURLExpire 生成指定有效期的临时下载 URL
func (s *OSSSigner) SignDownloadURLExpire(objectKey string, expireSeconds int) (string, error) {
	if s.AccessKey == "" || s.SecretKey == "" {
		return "", errors.New("OSS AccessKey/SecretKey 未配置")
	}
	if expireSeconds <= 0 {
		expireSeconds = 300
	}

	expire := time.Now().Unix() + int64(expireSeconds)
	host := s.host()

	// 构造待签名字符串: GET\n\n\n{expire}\n/{bucket}/{objectKey}
	resource := fmt.Sprintf("/%s/%s", s.Bucket, objectKey)
	stringToSign := fmt.Sprintf("GET\n\n\n%d\n%s", expire, resource)

	signature := s.sign(stringToSign)

	params := url.Values{}
	params.Set("OSSAccessKeyId", s.AccessKey)
	params.Set("Expires", strconv.FormatInt(expire, 10))
	params.Set("Signature", signature)

	return fmt.Sprintf("https://%s/%s?%s", host, url.PathEscape(objectKey), params.Encode()), nil
}

// host 返回访问域名
func (s *OSSSigner) host() string {
	if s.CDNDomain != "" {
		return strings.TrimPrefix(s.CDNDomain, "https://")
	}
	return fmt.Sprintf("%s.%s", s.Bucket, s.Endpoint)
}

// sign HMAC-SHA1 签名并 base64 编码
func (s *OSSSigner) sign(stringToSign string) string {
	h := hmac.New(sha1.New, []byte(s.SecretKey))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// SignUploadURL 生成临时上传 URL(PUT 方法)
func (s *OSSSigner) SignUploadURL(objectKey string, contentType string) (string, error) {
	if s.AccessKey == "" || s.SecretKey == "" {
		return "", errors.New("OSS AccessKey/SecretKey 未配置")
	}

	expire := time.Now().Unix() + int64(s.SignExpire)
	host := s.host()
	resource := fmt.Sprintf("/%s/%s", s.Bucket, objectKey)
	stringToSign := fmt.Sprintf("PUT\n%s\n\n%d\n%s", contentType, expire, resource)
	signature := s.sign(stringToSign)

	params := url.Values{}
	params.Set("OSSAccessKeyId", s.AccessKey)
	params.Set("Expires", strconv.FormatInt(expire, 10))
	params.Set("Signature", signature)

	return fmt.Sprintf("https://%s/%s?%s", host, url.PathEscape(objectKey), params.Encode()), nil
}

// PresignRequest 通用签名方法，供自定义请求使用
func (s *OSSSigner) PresignRequest(method, objectKey, contentType string, expireSeconds int) (*http.Request, error) {
	signedURL, err := s.SignDownloadURLExpire(objectKey, expireSeconds)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, signedURL, nil)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}
