package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	app "syairullahhajiumroh"
	"syairullahhajiumroh/internal/config"
	"syairullahhajiumroh/internal/handler"
	"syairullahhajiumroh/internal/middleware"
	"syairullahhajiumroh/internal/migration"
	"syairullahhajiumroh/internal/repository"

	_ "syairullahhajiumroh/docs"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// @title           Syairullah Haji & Umroh API
// @version         1.0
// @description     API untuk manajemen data jamaah haji dan umroh
// @host            localhost:8080
// @BasePath        /api
func main() {
	cfg := config.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	db := client.Database(cfg.MongoDB)

	// Run migrations
	if err := migration.RunMigrations(ctx, db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Ensure super admin from env
	adminRepo := repository.NewAdminRepository(db)
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}
	if err := adminRepo.EnsureSuperAdmin(ctx, cfg.AdminUsername, string(hashedPw)); err != nil {
		log.Fatalf("Failed to ensure super admin: %v", err)
	}
	log.Printf("Super admin ensured: %s", cfg.AdminUsername)

	// Create uploads directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup repository and handler
	jamaahRepo := repository.NewJamaahRepository(db)
	paketRepo := repository.NewPaketRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	jamaahHandler := handler.NewJamaahHandler(jamaahRepo, paketRepo, auditLogRepo, cfg.UploadDir)
	paketHandler := handler.NewPaketHandler(paketRepo, jamaahRepo, auditLogRepo)
	authHandler := handler.NewAuthHandler(adminRepo, cfg.JWTSecret)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)

	// Setup Gin router
	r := gin.Default()
	api := r.Group("/api")

	// Public routes
	authHandler.RegisterPublicRoutes(api)
	api.GET("/jamaah/:id/dokumen/:docType", jamaahHandler.GetDocument)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	authHandler.RegisterProtectedRoutes(protected)
	jamaahHandler.RegisterRoutes(protected)
	paketHandler.RegisterRoutes(protected)
	auditLogHandler.RegisterRoutes(protected)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve embedded React app
	distFS, err := fs.Sub(app.WebDist, "web/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	staticHandler := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try to serve static file
		f, err := distFS.Open(path[1:]) // strip leading /
		if err == nil {
			f.Close()
			staticHandler.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: serve index.html for all other routes
		c.Request.URL.Path = "/"
		staticHandler.ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Println("UI: http://localhost:" + cfg.AppPort)
	log.Println("Swagger UI: http://localhost:" + cfg.AppPort + "/swagger/index.html")
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
