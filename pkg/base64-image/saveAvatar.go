package avatar

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"strings"

	"github.com/minio/minio-go/v7"
)

func SaveAvatar(base64Str, filename string, minioClient *minio.Client) error {
	commaIndex := strings.Index(base64Str, ",")
	if commaIndex != -1 {
		base64Str = base64Str[commaIndex+1:]
	}

	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return fmt.Errorf("failed to decode base64 string: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, nil)
	if err != nil {
		return fmt.Errorf("failed to encode image to buffer: %w", err)
	}

	_, err = minioClient.PutObject(context.Background(), "avatars", filename, &buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	return nil
}
