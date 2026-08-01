package utils

import (
	"bytes"
	"errors"
	"strings"
)

// PDFMaxSize PDF 文件最大尺寸(10MB)
const PDFMaxSize = 10 * 1024 * 1024

// pdfMagicNumber PDF 文件头魔数
var pdfMagicNumber = []byte("%PDF-")

// ValidatePDF 校验上传文件是否为合法 PDF
// 1. 文件大小 ≤ 10MB
// 2. 文件头前 5 字节必须为 %PDF-
// 返回 (true, nil) 表示校验通过
func ValidatePDF(fileBytes []byte) (bool, error) {
	if len(fileBytes) == 0 {
		return false, errors.New("文件为空")
	}
	if len(fileBytes) > PDFMaxSize {
		return false, errors.New("文件大小超过 10MB 限制")
	}
	// 伪装检测:前 5 字节必须为 %PDF-
	if len(fileBytes) < 5 || !bytes.Equal(fileBytes[:5], pdfMagicNumber) {
		return false, errors.New("文件头魔数不是 %PDF-，疑似伪装文件")
	}
	return true, nil
}

// ValidatePDFFileExtension 校验文件扩展名必须为 .pdf
func ValidatePDFFileExtension(filename string) bool {
	if filename == "" {
		return false
	}
	return strings.HasSuffix(strings.ToLower(filename), ".pdf")
}
