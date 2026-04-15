package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	delivery "github.com/serhatkilbas/lms-poc/internal/delivery/http"
	"github.com/serhatkilbas/lms-poc/internal/delivery/http/middleware"
	"github.com/serhatkilbas/lms-poc/internal/domain"
	"github.com/serhatkilbas/lms-poc/internal/repository"
	"github.com/serhatkilbas/lms-poc/internal/usecase"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database Config
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local docker-compose
		dbURL = "postgres://lumina_user:lumina_password@localhost:5432/lumina_lms?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL")

	// Cloudflare R2 Config
	r2AccessKey := os.Getenv("R2_ACCESS_KEY_ID")
	r2SecretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	r2Endpoint := os.Getenv("R2_ENDPOINT")
	r2Bucket := os.Getenv("R2_BUCKET_NAME")

	// Initialize Repositories
	userRepo := repository.NewPostgresUserRepository(db)
	courseRepo := repository.NewPostgresCourseRepository(db)

	// Initialize Media Storage (R2)
	var mediaStorage domain.MediaStorage
	if r2AccessKey != "" && r2SecretKey != "" && r2Endpoint != "" && r2Bucket != "" {
		storage, err := repository.NewS3MediaStorage(r2AccessKey, r2SecretKey, r2Endpoint, r2Bucket)
		if err != nil {
			log.Fatalf("Failed to initialize R2 storage: %v", err)
		}
		mediaStorage = storage
		fmt.Println("R2 storage initialized")
	}

	// Use R2 for Video Repository too
	videoRepo := repository.NewR2VideoRepository(mediaStorage)

	// Initialize Transcoder
	videoTranscoder := repository.NewFFmpegTranscoder()

	// Initialize Progress components
	progressRepo := repository.NewMemoryProgressRepository()
	progressUsecase := usecase.NewProgressUsecase(progressRepo)
	progressHandler := delivery.NewProgressHandler(progressUsecase)

	// Initialize Auth components
	authUsecase := usecase.NewAuthUsecase(userRepo)
	authHandler := delivery.NewAuthHandler(authUsecase)

	// Initialize Course components
	courseUsecase := usecase.NewCourseUsecase(courseRepo, mediaStorage)
	courseHandler := delivery.NewCourseHandler(courseUsecase)

	// Initialize Document components
	pdfWatermarker := repository.NewPDFWatermarker()
	documentUsecase := usecase.NewDocumentUsecase(pdfWatermarker, mediaStorage)
	documentHandler := delivery.NewDocumentHandler(documentUsecase)

	// Initialize Usecase
	videoUsecase := usecase.NewVideoUsecase(videoRepo, mediaStorage, videoTranscoder)

	// Initialize Handler
	videoHandler := delivery.NewVideoHandler(videoUsecase)

	// Setup Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Setup CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "https://luminalms.netlify.app"} // Vite frontend origins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authHandler.Logout)
	r.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
	r.GET("/ws/progress", progressHandler.HandleWS)
	r.GET("/download/pdf", documentHandler.Download)

	// Protected Upload Routes
	uploadRoutes := r.Group("/upload")
	uploadRoutes.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(string(domain.RoleInstructor)))
	{
		uploadRoutes.POST("", videoHandler.UploadChunk)
		uploadRoutes.POST("/document", documentHandler.Upload)
	}

	// Course Routes (Protected)
	courseRoutes := r.Group("/courses")
	courseRoutes.Use(middleware.AuthMiddleware())
	{
		courseRoutes.GET("", courseHandler.List)
		courseRoutes.GET("/:id", courseHandler.Get)
		
		// Instructor only routes
		instructorMiddleware := middleware.RoleMiddleware(string(domain.RoleInstructor))
		courseRoutes.POST("", instructorMiddleware, courseHandler.Create)
		courseRoutes.PUT("/:id", instructorMiddleware, courseHandler.Update)
		courseRoutes.DELETE("/:id", instructorMiddleware, courseHandler.Delete)
		courseRoutes.POST("/:id/lessons", instructorMiddleware, courseHandler.AddLesson)
		courseRoutes.POST("/:id/lessons/reorder", instructorMiddleware, courseHandler.ReorderLessons)
		courseRoutes.PUT("/:id/lessons/:lessonId", instructorMiddleware, courseHandler.UpdateLesson)
		courseRoutes.DELETE("/:id/lessons/:lessonId", instructorMiddleware, courseHandler.DeleteLesson)
	}

	// Start Server
	fmt.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
