package usecase

import (
	"context"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type courseUsecase struct {
	repo         domain.CourseRepository
	mediaStorage domain.MediaStorage
}

func NewCourseUsecase(repo domain.CourseRepository, storage domain.MediaStorage) domain.CourseUsecase {
	return &courseUsecase{
		repo:         repo,
		mediaStorage: storage,
	}
}

func (u *courseUsecase) CreateCourse(ctx context.Context, course *domain.Course) error {
	if course.Status == "" {
		course.Status = domain.CourseStatusDraft
	}
	return u.repo.Create(ctx, course)
}

func (u *courseUsecase) GetCourse(ctx context.Context, id string) (*domain.Course, error) {
	course, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Generate Presigned URLs for private assets
	if u.mediaStorage != nil {
		for i := range course.Contents {
			item := &course.Contents[i]
			if (item.Type == domain.ContentTypeVideo || item.Type == domain.ContentTypeDocument) && item.ContentURL != "" {
				key := ""
				if strings.Contains(item.ContentURL, "/courses/") {
					parts := strings.Split(item.ContentURL, "/courses/")
					if len(parts) > 1 {
						key = "courses/" + parts[1]
					}
				}

				if key != "" {
					presigned, err := u.mediaStorage.GetPresignedURL(ctx, key, time.Hour*1)
					if err == nil {
						item.ContentURL = presigned
					}
				}
			}
		}
	}

	return course, nil
}

func (u *courseUsecase) GetAllCourses(ctx context.Context) ([]domain.Course, error) {
	return u.repo.List(ctx)
}

func (u *courseUsecase) UpdateCourse(ctx context.Context, course *domain.Course) error {
	if course.Status == "" {
		course.Status = domain.CourseStatusDraft
	}
	return u.repo.Update(ctx, course)
}

func (u *courseUsecase) DeleteCourse(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *courseUsecase) UpdateLesson(ctx context.Context, content *domain.CourseContent) error {
	return u.repo.UpdateContent(ctx, content)
}

func (u *courseUsecase) DeleteLesson(ctx context.Context, id string) error {
	return u.repo.DeleteContent(ctx, id)
}

func (u *courseUsecase) AddLessonToCourse(ctx context.Context, content *domain.CourseContent, file interface{}, fileName string) error {
	if file != nil && u.mediaStorage != nil {
		reader := file.(io.Reader)
		remotePath := "courses/" + content.CourseID + "/" + fileName

		contentType := mime.TypeByExtension(filepath.Ext(fileName))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		location, err := u.mediaStorage.UploadStream(ctx, reader, remotePath, contentType)
		if err != nil {
			return err
		}

		content.ContentURL = location
	}

	return u.repo.AddContent(ctx, content)
}
