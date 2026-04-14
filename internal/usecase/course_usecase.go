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
	// 1. Delete all assets from R2 first
	if u.mediaStorage != nil {
		// Delete course uploads
		_ = u.mediaStorage.DeleteDirectory(ctx, "courses/"+id+"/")
		// Delete potential streams (using courseID as part of uploadID logic if applicable)
		// For now, focusing on the main course directory
	}

	// 2. Delete from DB (metadata)
	return u.repo.Delete(ctx, id)
}

func (u *courseUsecase) UpdateLesson(ctx context.Context, content *domain.CourseContent) error {
	return u.repo.UpdateContent(ctx, content)
}

func (u *courseUsecase) DeleteLesson(ctx context.Context, id string) error {
	// 1. Get lesson details to find R2 path
	content, err := u.repo.GetContentByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Delete from R2 if applicable
	if u.mediaStorage != nil && content.ContentURL != "" {
		if content.Type == domain.ContentTypeVideo {
			// Extract uploadID from URL to delete HLS stream
			// Path is usually like streams/{uploadID}/index.m3u8
			if strings.Contains(content.ContentURL, "/streams/") {
				parts := strings.Split(content.ContentURL, "/streams/")
				if len(parts) > 1 {
					uploadID := strings.Split(parts[1], "/")[0]
					_ = u.mediaStorage.DeleteDirectory(ctx, "streams/"+uploadID+"/")
				}
			}
		} else if content.Type == domain.ContentTypeDocument {
			// Delete specific document file
			// Path is usually courses/{courseID}/{fileName}
			if strings.Contains(content.ContentURL, "/courses/") {
				parts := strings.Split(content.ContentURL, "/courses/")
				if len(parts) > 1 {
					key := "courses/" + parts[1]
					_ = u.mediaStorage.DeleteDirectory(ctx, key) // DeleteDirectory works for single file too
				}
			}
		}
	}

	// 3. Delete from DB
	return u.repo.DeleteContent(ctx, id)
}

func (u *courseUsecase) ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) error {
	return u.repo.BulkReorder(ctx, courseID, lessonIDs)
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
