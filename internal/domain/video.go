package domain

import (
	"context"
	"io"
)

type Chunk struct {
	Index       int
	TotalChunks int
	FileName    string
	UploadID    string
	Data        io.Reader
}

type VideoRepository interface {
	SaveChunk(ctx context.Context, chunk Chunk) error
	IsAllChunksUploaded(ctx context.Context, uploadID string, totalChunks int) (bool, error)
	AssembleChunks(ctx context.Context, uploadID, fileName string, totalChunks int) (string, error)
	CleanupChunks(ctx context.Context, uploadID string) error
}

type VideoUsecase interface {
	ProcessChunk(ctx context.Context, chunk Chunk) (bool, error) // Returns true if complete
}

type MediaStorage interface {
	UploadFile(ctx context.Context, filePath, fileName string) (string, error)
	UploadStream(ctx context.Context, reader io.Reader, fileName, contentType string) (string, error)
	UploadDirectory(ctx context.Context, localDir, remotePrefix string) error
}


type Transcoder interface {
	TranscodeToHLS(ctx context.Context, inputPath, outputDir string) (string, error) // Returns path to master playlist
}


