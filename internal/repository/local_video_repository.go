package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type localVideoRepository struct {
	uploadDir string
	mutex     sync.Mutex
}

func NewLocalVideoRepository(uploadDir string) domain.VideoRepository {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &localVideoRepository{uploadDir: uploadDir}
}

func (r *localVideoRepository) SaveChunk(ctx context.Context, chunk domain.Chunk) error {
	tempDir := filepath.Join(r.uploadDir, chunk.UploadID)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return err
		}
	}

	chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk-%d", chunk.Index))
	out, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, chunk.Data)
	return err
}

func (r *localVideoRepository) IsAllChunksUploaded(ctx context.Context, uploadID string, totalChunks int) (bool, error) {
	tempDir := filepath.Join(r.uploadDir, uploadID)
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return false, err
	}
	return len(files) == totalChunks, nil
}

func (r *localVideoRepository) AssembleChunks(ctx context.Context, uploadID, fileName string, totalChunks int) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tempDir := filepath.Join(r.uploadDir, uploadID)
	finalPath := filepath.Join(r.uploadDir, fileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer finalFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk-%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return "", err
		}
	}

	return finalPath, nil
}

func (r *localVideoRepository) CleanupChunks(ctx context.Context, uploadID string) error {
	tempDir := filepath.Join(r.uploadDir, uploadID)
	return os.RemoveAll(tempDir)
}
