package main

import (
	"context"
	"log"
	"os"
	"time"

	"syairullahhajiumroh/internal/config"
	"syairullahhajiumroh/internal/handler"
	"syairullahhajiumroh/internal/migration"
	"syairullahhajiumroh/internal/repository"

	_ "syairullahhajiumroh/docs"

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

	// Create uploads directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup repository and handler
	jamaahRepo := repository.NewJamaahRepository(db)
	jamaahHandler := handler.NewJamaahHandler(jamaahRepo, cfg.UploadDir)

	// Setup Gin router
	r := gin.Default()
	api := r.Group("/api")
	jamaahHandler.RegisterRoutes(api)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Println("Swagger UI: http://localhost:" + cfg.AppPort + "/swagger/index.html")
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
