package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type videoUsecase struct {
	videoRepo    domain.VideoRepository
	mediaStorage domain.MediaStorage
	transcoder   domain.Transcoder
}

func NewVideoUsecase(repo domain.VideoRepository, storage domain.MediaStorage, transcoder domain.Transcoder) domain.VideoUsecase {
	return &videoUsecase{
		videoRepo:    repo,
		mediaStorage: storage,
		transcoder:   transcoder,
	}
}

func (u *videoUsecase) ProcessChunk(ctx context.Context, chunk domain.Chunk) (bool, error) {
	if err := u.videoRepo.SaveChunk(ctx, chunk); err != nil {
		return false, err
	}

	isComplete, err := u.videoRepo.IsAllChunksUploaded(ctx, chunk.UploadID, chunk.TotalChunks)
	if err != nil {
		return false, err
	}

	if isComplete {
		// Assemble
		localPath, err := u.videoRepo.AssembleChunks(ctx, chunk.UploadID, chunk.FileName, chunk.TotalChunks)
		if err != nil {
			return true, err
		}

		// Transcode to HLS
		if u.transcoder != nil {
			outputDir := filepath.Join(filepath.Dir(localPath), chunk.UploadID+"_hls")
			_, err = u.transcoder.TranscodeToHLS(ctx, localPath, outputDir)
			if err != nil {
				fmt.Printf("Failed to transcode: %v\n", err)
				return true, err
			}
			fmt.Printf("Successfully transcoded %s to HLS\n", chunk.FileName)

			// Upload HLS directory to R2
			if u.mediaStorage != nil {
				remotePrefix := "videos/" + chunk.UploadID
				err = u.mediaStorage.UploadDirectory(ctx, outputDir, remotePrefix)
				if err != nil {
					fmt.Printf("Failed to upload HLS directory: %v\n", err)
					return true, err
				}
				fmt.Printf("Successfully uploaded HLS for %s to R2\n", chunk.FileName)
			}

			// Cleanup transcoded files
			os.RemoveAll(outputDir)
		} else if u.mediaStorage != nil {
			// If no transcoder, just upload the raw file
			_, err = u.mediaStorage.UploadFile(ctx, localPath, chunk.FileName)
			if err != nil {
				fmt.Printf("Failed to upload to storage: %v\n", err)
				return true, err
			}
			fmt.Printf("Successfully uploaded %s to R2\n", chunk.FileName)
		}

		// Cleanup chunks and the assembled local file
		_ = u.videoRepo.CleanupChunks(ctx, chunk.UploadID)
		os.Remove(localPath)
		return true, nil
	}

	return false, nil
}
