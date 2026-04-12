package domain

import (
	"context"
	"io"
)

type Watermarker interface {
	WatermarkPDF(ctx context.Context, input io.ReadSeeker, text string) (io.Reader, error)
}

type DocumentUsecase interface {
	Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (string, error)
	GetWatermarkedPDF(ctx context.Context, inputPath, userID string) (io.Reader, error)
}
