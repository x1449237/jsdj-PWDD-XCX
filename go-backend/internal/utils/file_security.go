package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jisan/e-sports-platform/internal/model"
	"gorm.io/gorm"
)

// 文件安全相关工具：水印、AES-256 加密、SHA-256 哈希、上链存证

// 水印魔法标记，用于在二进制末尾区分水印段
const watermarkMagic = "JISANWM:"

// EmbedWatermark 在文件二进制末尾追加水印数据
// 格式: <原始内容><magic><watermarkText 长度(4字节大端)><watermarkText>
func EmbedWatermark(fileBytes []byte, watermarkText string) []byte {
	if len(fileBytes) == 0 {
		return fileBytes
	}
	wm := []byte(watermarkText)
	out := make([]byte, 0, len(fileBytes)+len(watermarkMagic)+4+len(wm))
	out = append(out, fileBytes...)
	out = append(out, []byte(watermarkMagic)...)
	// 4 字节大端长度
	l := uint32(len(wm))
	out = append(out, byte(l>>24), byte(l>>16), byte(l>>8), byte(l))
	out = append(out, wm...)
	return out
}

// VerifyWatermark 校验文件末尾是否携带水印，返回 (水印文本, 是否命中)
func VerifyWatermark(fileBytes []byte) (string, bool) {
	magicLen := len(watermarkMagic)
	if len(fileBytes) < magicLen+4 {
		return "", false
	}
	// 从后向前查找 magic 标记
	maxStart := len(fileBytes) - magicLen - 4
	for i := maxStart; i >= 0; i-- {
		if string(fileBytes[i:i+magicLen]) != watermarkMagic {
			continue
		}
		wmLenStart := i + magicLen
		if wmLenStart+4 > len(fileBytes) {
			return "", false
		}
		l := uint32(fileBytes[wmLenStart])<<24 |
			uint32(fileBytes[wmLenStart+1])<<16 |
			uint32(fileBytes[wmLenStart+2])<<8 |
			uint32(fileBytes[wmLenStart+3])
		wmStart := wmLenStart + 4
		if wmStart+int(l) > len(fileBytes) {
			return "", false
		}
		return string(fileBytes[wmStart : wmStart+int(l)]), true
	}
	return "", false
}

// EncryptFile 使用 AES-256-GCM 加密文件字节流，返回 base64 编码字符串
// key 必须为 32 字节
func EncryptFile(fileBytes []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("AES-256 密钥长度必须为 32 字节")
	}
	return AESEncrypt(string(fileBytes), key)
}

// DecryptFile 解密 AES-256-GCM 加密的 base64 密文
func DecryptFile(encoded string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("AES-256 密钥长度必须为 32 字节")
	}
	plain, err := AESDecrypt(encoded, key)
	if err != nil {
		return nil, err
	}
	return []byte(plain), nil
}

// GenerateFileHash 计算文件 SHA-256 哈希(十六进制字符串)
func GenerateFileHash(fileBytes []byte) string {
	sum := sha256.Sum256(fileBytes)
	return hex.EncodeToString(sum[:])
}

// defaultDB 默认 DB 句柄，由 SetFileSecurityDB 注入，供 UploadToBlockchain 兜底使用
var defaultDB *gorm.DB

// SetFileSecurityDB 注入默认 DB 句柄(供 UploadToBlockchain 使用)
func SetFileSecurityDB(db *gorm.DB) {
	defaultDB = db
}

// UploadToBlockchain 模拟上链存证:写入 file_blockchain_records 表，返回上链交易ID
// hash 为文件 SHA-256 哈希；metadata 携带 file_type/ref_type/ref_id/oss_url/watermark_text
func UploadToBlockchain(hash string, metadata map[string]string) (string, error) {
	if defaultDB == nil {
		return "", errors.New("DB 未注入，无法上链存证")
	}
	if hash == "" {
		return "", errors.New("文件哈希不能为空")
	}
	txID := fmt.Sprintf("bc-%s-%d", hash[:minInt(8, len(hash))], time.Now().UnixNano())
	refID := int64(0)
	for _, ch := range metadata["ref_id"] {
		if ch < '0' || ch > '9' {
			break
		}
		refID = refID*10 + int64(ch-'0')
	}
	rec := &model.FileBlockchainRecord{
		FileHash:       hash,
		FileType:       metadata["file_type"],
		RefType:        metadata["ref_type"],
		RefID:          refID,
		OSSUrl:         metadata["oss_url"],
		WatermarkText:  metadata["watermark_text"],
		BlockchainTxID: txID,
		CreatedAt:      ptrTime(time.Now()),
	}
	if err := defaultDB.Create(rec).Error; err != nil {
		return "", err
	}
	return txID, nil
}

// ptrTime 返回给定 time.Time 的指针
func ptrTime(t time.Time) *time.Time {
	return &t
}

// minInt 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
