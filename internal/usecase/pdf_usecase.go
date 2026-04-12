package usecase

import (
	"context"
	"io"
	"os"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type documentUsecase struct {
	watermarker  domain.Watermarker
	mediaStorage domain.MediaStorage
}

func NewDocumentUsecase(watermarker domain.Watermarker, storage domain.MediaStorage) domain.DocumentUsecase {
	return &documentUsecase{
		watermarker:  watermarker,
		mediaStorage: storage,
	}
}

func (u *documentUsecase) Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (string, error) {
	if u.mediaStorage == nil {
		return "", os.ErrNotExist // Or handle differently
	}
	
	remotePrefix := "documents/" + fileName
	return u.mediaStorage.UploadStream(ctx, reader, remotePrefix, contentType)
}

func (u *documentUsecase) GetWatermarkedPDF(ctx context.Context, inputPath, userID string) (io.Reader, error) {
	// Open existing PDF
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Apply watermark
	return u.watermarker.WatermarkPDF(ctx, file, userID)
}
