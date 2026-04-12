package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type memoryProgressRepository struct {
	progress map[string]float64 // key: userID:videoID
	mutex    sync.RWMutex
}

func NewMemoryProgressRepository() domain.ProgressRepository {
	return &memoryProgressRepository{
		progress: make(map[string]float64),
	}
}

func (r *memoryProgressRepository) SaveProgress(ctx context.Context, p domain.UserProgress) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", p.UserID, p.VideoID)
	r.progress[key] = p.Progress
	return nil
}

func (r *memoryProgressRepository) GetProgress(ctx context.Context, userID, videoID string) (float64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, videoID)
	if p, ok := r.progress[key]; ok {
		return p, nil
	}
	return 0, nil
}
