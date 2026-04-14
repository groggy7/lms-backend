package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type r2VideoRepository struct {
	storage domain.MediaStorage
	// Track chunks in memory for this POC (In production, use Redis)
	chunks map[string]map[int]string
	mu     sync.RWMutex
}

func NewR2VideoRepository(storage domain.MediaStorage) domain.VideoRepository {
	return &r2VideoRepository{
		storage: storage,
		chunks:  make(map[string]map[int]string),
	}
}

func (r *r2VideoRepository) SaveChunk(ctx context.Context, chunk domain.Chunk) error {
	r.mu.Lock()
	if _, ok := r.chunks[chunk.UploadID]; !ok {
		r.chunks[chunk.UploadID] = make(map[int]string)
	}
	r.mu.Unlock()

	tempDir := filepath.Join("./temp_chunks", chunk.UploadID)
	os.MkdirAll(tempDir, 0755)

	chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", chunk.Index))
	f, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, chunk.Data); err != nil {
		return err
	}

	r.mu.Lock()
	r.chunks[chunk.UploadID][chunk.Index] = chunkPath
	r.mu.Unlock()

	return nil
}

func (r *r2VideoRepository) IsAllChunksUploaded(ctx context.Context, uploadID string, totalChunks int) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if chunkMap, ok := r.chunks[uploadID]; ok {
		return len(chunkMap) == totalChunks, nil
	}
	return false, nil
}

func (r *r2VideoRepository) AssembleChunks(ctx context.Context, uploadID, fileName string, totalChunks int) (string, error) {
	r.mu.RLock()
	chunkMap := r.chunks[uploadID]
	r.mu.RUnlock()

	if len(chunkMap) != totalChunks {
		return "", fmt.Errorf("not all chunks uploaded")
	}

	// Create finalized local file
	finalPath := filepath.Join("./uploads", fileName)
	os.MkdirAll("./uploads", 0755)
	
	f, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Sort indices to ensure correct order
	indices := make([]int, 0, len(chunkMap))
	for idx := range chunkMap {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	for _, idx := range indices {
		chunkFile, err := os.Open(chunkMap[idx])
		if err != nil {
			return "", err
		}
		io.Copy(f, chunkFile)
		chunkFile.Close()
	}

	// NOTE: We return the LOCAL path now, so the Usecase can transcode it before uploading to R2
	return finalPath, nil
}

func (r *r2VideoRepository) CleanupChunks(ctx context.Context, uploadID string) error {
	r.mu.Lock()
	delete(r.chunks, uploadID)
	r.mu.Unlock()
	
	return os.RemoveAll(filepath.Join("./temp_chunks", uploadID))
}
