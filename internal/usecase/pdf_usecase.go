package usecase

import (
	"context"
	"io"
	"os"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type pdfUsecase struct {
	watermarker domain.Watermarker
}

func NewPDFUsecase(watermarker domain.Watermarker) domain.PDFUsecase {
	return &pdfUsecase{watermarker: watermarker}
}

func (u *pdfUsecase) GetWatermarkedPDF(ctx context.Context, inputPath, userID string) (io.Reader, error) {
	// Open existing PDF
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Apply watermark
	return u.watermarker.WatermarkPDF(ctx, file, userID)
}
