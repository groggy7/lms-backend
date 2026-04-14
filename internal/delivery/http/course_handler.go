package http

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type CourseHandler struct {
	usecase domain.CourseUsecase
}

func NewCourseHandler(usecase domain.CourseUsecase) *CourseHandler {
	return &CourseHandler{usecase: usecase}
}

func (h *CourseHandler) Create(c *gin.Context) {
	var course domain.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Instructor identity not found"})
		return
	}

	course.InstructorID = userID.(string)

	if err := h.usecase.CreateCourse(c.Request.Context(), &course); err != nil {
		fmt.Printf("[DEBUG] Course creation failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize course", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) List(c *gin.Context) {
	courses, err := h.usecase.GetAllCourses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) Get(c *gin.Context) {
	id := c.Param("id")
	course, err := h.usecase.GetCourse(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var course domain.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	course.ID = id

	if err := h.usecase.UpdateCourse(c.Request.Context(), &course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	fmt.Printf("[DEBUG] Attempting to delete course: %s\n", id)
	if err := h.usecase.DeleteCourse(c.Request.Context(), id); err != nil {
		fmt.Printf("[DEBUG] Deletion failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

func (h *CourseHandler) AddLesson(c *gin.Context) {
	courseID := c.Param("id")
	
	title := c.PostForm("title")
	contentType := c.PostForm("type")
	contentBody := c.PostForm("contentBody")
	orderIndexStr := c.PostForm("orderIndex")
	orderIndex, _ := strconv.Atoi(orderIndexStr)

	content := &domain.CourseContent{
		CourseID:    courseID,
		Title:       title,
		Type:        domain.ContentType(contentType),
		ContentBody: contentBody,
		OrderIndex:  orderIndex,
	}

	fileHeader, err := c.FormFile("asset")
	var fileStream interface{} = nil
	var fileName string = ""
	
	if err == nil {
		f, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open asset"})
			return
		}
		defer f.Close()
		fileStream = f
		
		// Use sanitized title as the filename
		ext := filepath.Ext(fileHeader.Filename)
		sanitizedTitle := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '-'
		}, title)
		fileName = strings.ToLower(sanitizedTitle) + ext
	}

	if err := h.usecase.AddLessonToCourse(c.Request.Context(), content, fileStream, fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, content)
}

func (h *CourseHandler) UpdateLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var content domain.CourseContent
	if err := c.ShouldBindJSON(&content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	content.ID = lessonID

	if err := h.usecase.UpdateLesson(c.Request.Context(), &content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

func (h *CourseHandler) DeleteLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	if err := h.usecase.DeleteLesson(c.Request.Context(), lessonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
}

func (h *CourseHandler) ReorderLessons(c *gin.Context) {
	courseID := c.Param("id")
	var req struct {
		LessonIDs []string `json:"lessonIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ReorderLessons(c.Request.Context(), courseID, req.LessonIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lessons reordered successfully"})
}
