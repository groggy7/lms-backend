package domain

import (
	"context"
)

type UserProgress struct {
	UserID   string  `json:"userId"`
	VideoID  string  `json:"videoId"`
	Progress float64 `json:"progress"` // Current time in seconds
}

type ProgressRepository interface {
	SaveProgress(ctx context.Context, progress UserProgress) error
	GetProgress(ctx context.Context, userID, videoID string) (float64, error)
}

type ProgressUsecase interface {
	UpdateProgress(ctx context.Context, progress UserProgress) error
	GetLatestProgress(ctx context.Context, userID, videoID string) (float64, error)
}
