package encoder

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"

	_ "golang.org/x/image/webp"
)

func EncodeImageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	err = png.Encode(&buf, rgba)
	if err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	base64Encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	mimeType := "image/png"

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Encoded), nil
}
