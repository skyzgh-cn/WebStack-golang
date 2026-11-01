package admin

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

// validateImageContent 通过读取文件头部字节判断内容是否确为受支持的图片，
// 防止攻击者将可执行脚本伪装成图片上传（例如 logo.php）。
func validateImageContent(file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return errors.New("无法读取上传文件")
	}
	defer src.Close()

	// 读取文件头用于识别类型
	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return errors.New("无法读取上传文件内容")
	}
	header = header[:n]

	if len(header) == 0 {
		return errors.New("上传文件为空")
	}

	contentType := http.DetectContentType(header)

	// DetectContentType 对 SVG 会返回 text/xml 或 text/plain，
	// 因此这里只接受明确的图片类型
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp":
		return nil
	}

	// 额外识别 DetectContentType 未覆盖的 PNG/JPEG 变体
	if isKnownImageSignature(header) {
		return nil
	}

	return errors.New("文件内容不是有效的图片")
}

// isKnownImageSignature 通过魔数判断文件是否为受支持的图片格式。
func isKnownImageSignature(header []byte) bool {
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(header) >= 8 && bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return true
	}
	// JPEG: FF D8 FF
	if len(header) >= 3 && bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}) {
		return true
	}
	// GIF: "GIF87a" 或 "GIF89a"
	if len(header) >= 6 && (bytes.HasPrefix(header, []byte("GIF87a")) || bytes.HasPrefix(header, []byte("GIF89a"))) {
		return true
	}
	// RIFF....WEBP
	if len(header) >= 12 && bytes.HasPrefix(header, []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")) {
		return true
	}
	// BMP: "BM"
	if len(header) >= 2 && bytes.HasPrefix(header, []byte("BM")) {
		return true
	}
	return false
}
