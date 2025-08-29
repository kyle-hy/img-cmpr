package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	// 用于支持 WebP 格式解码
)

const (
	inputDir    = "images"
	outputDir   = "img_cmp"
	maxSizeKB   = 480
	minQuality  = 10
	initialQual = 100
	qualityStep = 5
)

var supportedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func main() {
	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		outDir := filepath.Join(outputDir, filepath.Dir(relPath))
		os.MkdirAll(outDir, os.ModePerm)
		outPath := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(path), ext)+".jpg")

		success := compressImage(path, outPath)
		if success {
			fmt.Printf("✅ %s → %s\n", path, outPath)
		} else {
			fmt.Printf("❌ %s (无法压缩到目标大小)\n", path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("发生错误:", err)
	}
}

func compressImage(inputPath, outputPath string) bool {
	file, err := os.Open(inputPath)
	if err != nil {
		return false
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return false
	}

	// 不支持的格式
	if img == nil {
		return false
	}

	// PNG/WebP 转换为 RGB
	if format == "png" || format == "webp" {
		img = convertToRGB(img)
	}

	quality := initialQual
	for quality >= minQuality {
		var buf bytes.Buffer
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return false
		}

		sizeKB := float64(buf.Len()) / 1024
		if sizeKB <= maxSizeKB {
			return os.WriteFile(outputPath, buf.Bytes(), 0644) == nil
		}
		quality -= qualityStep
	}

	return false
}

// convertToRGB 确保所有图像格式都能以 RGB 保存
func convertToRGB(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	return dst
}
