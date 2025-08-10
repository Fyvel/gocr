package image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func (ip *ImageProcessor) EnhanceQuality(path string) (string, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening image %s: %w", path, err)
	}

	// Resize if too small
	bounds := img.Bounds()
	if bounds.Dx() < 300 || bounds.Dy() < 300 {
		img = imaging.Resize(img, bounds.Dx()*2, bounds.Dy()*2, imaging.Lanczos)
	}

	gray := imaging.Grayscale(img)
	contrast := imaging.AdjustContrast(gray, 10)
	sharp := imaging.Sharpen(contrast, 1.1)

	extension := filepath.Ext(path)
	localPath := path[:len(path)-len(extension)]
	tempPath := localPath + "_processed" + extension
	if err := imaging.Save(sharp, tempPath); err != nil {
		// if err := imaging.Save(img, tempPath); err != nil {
		return "", fmt.Errorf("saving processed image: %w", err)
	}

	return tempPath, nil
}

func (ip *ImageProcessor) Cleanup(filePath string) error {
	return os.Remove(filePath)
}
