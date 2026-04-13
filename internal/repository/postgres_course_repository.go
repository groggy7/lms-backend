package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type postgresCourseRepository struct {
	db *sql.DB
}

func NewPostgresCourseRepository(db *sql.DB) domain.CourseRepository {
	return &postgresCourseRepository{db: db}
}

func (r *postgresCourseRepository) Create(ctx context.Context, course *domain.Course) error {
	course.ID = fmt.Sprintf("crs_%d", rand.Intn(1000000))
	query := `
		INSERT INTO courses (id, title, description, instructor_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		course.ID,
		course.Title,
		course.Description,
		course.InstructorID,
		course.Status,
	).Scan(&course.CreatedAt, &course.UpdatedAt)
}

func (r *postgresCourseRepository) GetByID(ctx context.Context, id string) (*domain.Course, error) {
	query := `SELECT id, title, description, instructor_id, status, created_at, updated_at FROM courses WHERE id = $1`

	course := &domain.Course{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&course.ID,
		&course.Title,
		&course.Description,
		&course.InstructorID,
		&course.Status,
		&course.CreatedAt,
		&course.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	contents, err := r.GetContentByCourse(ctx, id)
	if err == nil {
		course.Contents = contents
	}

	return course, nil
}

func (r *postgresCourseRepository) List(ctx context.Context) ([]domain.Course, error) {
	query := `SELECT id, title, description, instructor_id, status, created_at, updated_at FROM courses ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.InstructorID, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, nil
}

func (r *postgresCourseRepository) Update(ctx context.Context, course *domain.Course) error {
	query := `
		UPDATE courses 
		SET title = $1, description = $2, status = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, course.Title, course.Description, course.Status, course.ID)
	return err
}

func (r *postgresCourseRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresCourseRepository) AddContent(ctx context.Context, content *domain.CourseContent) error {
	content.ID = fmt.Sprintf("cnt_%d", rand.Intn(1000000))
	query := `
		INSERT INTO course_contents (id, course_id, title, type, content_url, content_body, order_index, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		content.ID,
		content.CourseID,
		content.Title,
		content.Type,
		content.ContentURL,
		content.ContentBody,
		content.OrderIndex,
	).Scan(&content.CreatedAt)
}

func (r *postgresCourseRepository) UpdateContent(ctx context.Context, content *domain.CourseContent) error {
	query := `
		UPDATE course_contents 
		SET title = $1, content_body = $2, order_index = $3 
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, content.Title, content.ContentBody, content.OrderIndex, content.ID)
	return err
}

func (r *postgresCourseRepository) DeleteContent(ctx context.Context, id string) error {
	fmt.Printf("[DEBUG] Repository: Attempting to delete content id: %s\n", id)
	query := `DELETE FROM course_contents WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	fmt.Printf("[DEBUG] Repository: Deletion successful, rows affected: %d\n", rows)
	return nil
}

func (r *postgresCourseRepository) BulkReorder(ctx context.Context, courseID string, lessonIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range lessonIDs {
		query := `UPDATE course_contents SET order_index = $1 WHERE id = $2 AND course_id = $3`
		if _, err := tx.ExecContext(ctx, query, i, id, courseID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresCourseRepository) GetContentByCourse(ctx context.Context, courseID string) ([]domain.CourseContent, error) {
	query := `
		SELECT id, course_id, title, type, content_url, content_body, order_index, created_at 
		FROM course_contents 
		WHERE course_id = $1 
		ORDER BY order_index ASC
	`

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []domain.CourseContent
	for rows.Next() {
		var c domain.CourseContent
		if err := rows.Scan(&c.ID, &c.CourseID, &c.Title, &c.Type, &c.ContentURL, &c.ContentBody, &c.OrderIndex, &c.CreatedAt); err != nil {
			return nil, err
		}
		contents = append(contents, c)
	}
	return contents, nil
}
