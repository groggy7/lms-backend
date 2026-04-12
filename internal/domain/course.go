package domain

import (
	"context"
	"time"
)

type ContentType string

const (
	ContentTypeVideo    ContentType = "video"
	ContentTypeDocument ContentType = "document"
	ContentTypeText     ContentType = "text"
)

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
)

type Course struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	InstructorID string         `json:"instructorId"`
	Status       CourseStatus   `json:"status"`
	Contents     []CourseContent `json:"contents,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

type CourseContent struct {
	ID          string      `json:"id"`
	CourseID    string      `json:"courseId"`
	Title       string      `json:"title"`
	Type        ContentType `json:"type"`
	ContentURL  string      `json:"contentUrl,omitempty"`
	ContentBody string      `json:"contentBody,omitempty"`
	OrderIndex  int         `json:"orderIndex"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type CourseRepository interface {
	Create(ctx context.Context, course *Course) error
	GetByID(ctx context.Context, id string) (*Course, error)
	List(ctx context.Context) ([]Course, error)
	Update(ctx context.Context, course *Course) error
	Delete(ctx context.Context, id string) error
	
	// Content operations
	AddContent(ctx context.Context, content *CourseContent) error
	UpdateContent(ctx context.Context, content *CourseContent) error
	DeleteContent(ctx context.Context, id string) error
	GetContentByCourse(ctx context.Context, courseID string) ([]CourseContent, error)
}

type CourseUsecase interface {
	CreateCourse(ctx context.Context, course *Course) error
	GetCourse(ctx context.Context, id string) (*Course, error)
	GetAllCourses(ctx context.Context) ([]Course, error)
	UpdateCourse(ctx context.Context, course *Course) error
	DeleteCourse(ctx context.Context, id string) error
	AddLessonToCourse(ctx context.Context, content *CourseContent, file interface{}, fileName string) error
	UpdateLesson(ctx context.Context, content *CourseContent) error
	DeleteLesson(ctx context.Context, id string) error
}
