package domain

import (
	"context"
	"io"
)

type Watermarker interface {
	WatermarkPDF(ctx context.Context, input io.ReadSeeker, text string) (io.Reader, error)
}

type PDFUsecase interface {
	GetWatermarkedPDF(ctx context.Context, inputPath, userID string) (io.Reader, error)
}
