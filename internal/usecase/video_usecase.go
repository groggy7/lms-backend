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

		// ASYNCHRONOUS HIGH-PERFORMANCE PROCESSING
		// We launch a background routine so the API response isn't blocked by FFmpeg
		go func() {
			bgCtx := context.Background()
			
			// 1. Transcode to HLS
			if u.transcoder != nil {
				outputDir := filepath.Join("./hls_output", chunk.UploadID)
				_, err = u.transcoder.TranscodeToHLS(bgCtx, localPath, outputDir)
				if err != nil {
					fmt.Printf("[ERROR] Failed to transcode %s: %v\n", chunk.FileName, err)
					return
				}
				fmt.Printf("[INFO] Successfully transcoded %s to HLS\n", chunk.FileName)

				// 2. Upload HLS segments and playlist to R2
				if u.mediaStorage != nil {
					remotePrefix := "streams/" + chunk.UploadID
					err = u.mediaStorage.UploadDirectory(bgCtx, outputDir, remotePrefix)
					if err != nil {
						fmt.Printf("[ERROR] Failed to upload HLS to R2: %v\n", err)
						return
					}
					fmt.Printf("[INFO] HLS segments for %s deployed to R2\n", chunk.FileName)
				}

				// 3. Cleanup local HLS files
				os.RemoveAll(outputDir)
			} else if u.mediaStorage != nil {
				// Fallback: If no transcoder, just upload raw MP4 to R2
				remotePath := "uploads/" + chunk.FileName
				_, err = u.mediaStorage.UploadFile(bgCtx, localPath, remotePath)
				if err != nil {
					fmt.Printf("[ERROR] Failed fallback upload to R2: %v\n", err)
					return
				}
			}

			// 4. Cleanup raw assembled file and chunks
			_ = u.videoRepo.CleanupChunks(bgCtx, chunk.UploadID)
			os.Remove(localPath)
		}()

		return true, nil
	}

	return false, nil
}
