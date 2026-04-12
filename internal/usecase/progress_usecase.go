package usecase

import (
	"context"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type progressUsecase struct {
	repo domain.ProgressRepository
}

func NewProgressUsecase(repo domain.ProgressRepository) domain.ProgressUsecase {
	return &progressUsecase{repo: repo}
}

func (u *progressUsecase) UpdateProgress(ctx context.Context, progress domain.UserProgress) error {
	return u.repo.SaveProgress(ctx, progress)
}

func (u *progressUsecase) GetLatestProgress(ctx context.Context, userID, videoID string) (float64, error) {
	return u.repo.GetProgress(ctx, userID, videoID)
}
