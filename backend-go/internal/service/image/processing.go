package image

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

// Processor 图片处理器，负责缩放和缩略图生成
type Processor struct {
	MaxWidth  int
	MaxHeight int
}

// ProcessedImage 处理后的图片结果
type ProcessedImage struct {
	Data       []byte // 处理后的原图数据
	Width      uint
	Height     uint
	Thumbnail  []byte // 缩略图数据
	ThumbWidth uint
	ThumbHeight uint
}

// NewProcessor 创建图片处理器
func NewProcessor(maxWidth, maxHeight int) *Processor {
	if maxWidth <= 0 {
		maxWidth = 4096
	}
	if maxHeight <= 0 {
		maxHeight = 4096
	}
	return &Processor{MaxWidth: maxWidth, MaxHeight: maxHeight}
}

// Process 处理图片：超标缩放 + 生成缩略图
func (p *Processor) Process(data []byte) (*ProcessedImage, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}

	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// 如果图片尺寸超出限制则等比缩放
	if width > uint(p.MaxWidth) || height > uint(p.MaxHeight) {
		img = imaging.Fit(img, p.MaxWidth, p.MaxHeight, imaging.Lanczos)
		bounds = img.Bounds()
	}

	// 编码处理后的图片
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return nil, fmt.Errorf("图片编码失败: %w", err)
	}

	// 生成缩略图 (最大 400px)
	thumb := imaging.Thumbnail(img, 400, 400, imaging.Lanczos)
	thumbBounds := thumb.Bounds()

	var thumbBuf bytes.Buffer
	if err := jpeg.Encode(&thumbBuf, thumb, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("缩略图生成失败: %w", err)
	}

	return &ProcessedImage{
		Data:        buf.Bytes(),
		Width:       uint(bounds.Dx()),
		Height:      uint(bounds.Dy()),
		Thumbnail:   thumbBuf.Bytes(),
		ThumbWidth:  uint(thumbBounds.Dx()),
		ThumbHeight: uint(thumbBounds.Dy()),
	}, nil
}
