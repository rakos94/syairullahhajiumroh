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

	// Refuse to run in production with published default secrets.
	fatal, warnings := cfg.SecretProblems()
	for _, w := range warnings {
		log.Printf("PERINGATAN KEAMANAN: %s", w)
	}
	for _, f := range fatal {
		if cfg.Production {
			log.Printf("KEAMANAN: %s", f)
		} else {
			log.Printf("PERINGATAN KEAMANAN (dev, diizinkan): %s", f)
		}
	}
	if cfg.Production && len(fatal) > 0 {
		log.Fatal("Aplikasi tidak dijalankan karena konfigurasi keamanan di atas. Perbaiki .env lalu jalankan ulang.")
	}

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

	// Super admin: created from ADMIN_USERNAME/ADMIN_PASSWORD only when none
	// exists yet. Afterwards the password is managed in the app (menu Admin);
	// ADMIN_RESET_PASSWORD=true resets it from .env for account recovery.
	adminRepo := repository.NewAdminRepository(db)
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}
	if cfg.Production && cfg.AdminPassword == config.DefaultAdminPassword {
		if n, _ := adminRepo.Count(ctx); n == 0 || cfg.AdminResetPassword {
			log.Fatal("KEAMANAN: ADMIN_PASSWORD masih default (admin123) untuk membuat/mereset super admin. Ganti di .env lalu jalankan ulang.")
		}
	}
	created, reset, err := adminRepo.EnsureSuperAdmin(ctx, cfg.AdminUsername, string(hashedPw), cfg.AdminResetPassword)
	if err != nil {
		log.Fatalf("Failed to ensure super admin: %v", err)
	}
	switch {
	case created:
		log.Printf("Super admin dibuat: %s", cfg.AdminUsername)
	case reset:
		log.Printf("PERINGATAN: password super admin direset dari .env (ADMIN_RESET_PASSWORD=true). Hapus pengaturan itu setelah login.")
	default:
		log.Printf("Super admin sudah ada; password dikelola di menu Admin (bukan dari .env)")
	}

	// Create uploads directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup repository and handler
	jamaahRepo := repository.NewJamaahRepository(db)
	paketRepo := repository.NewPaketRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	jamaahHandler := handler.NewJamaahHandler(jamaahRepo, paketRepo, auditLogRepo, cfg.UploadDir, cfg.JWTSecret)
	paketHandler := handler.NewPaketHandler(paketRepo, jamaahRepo, auditLogRepo)
	authHandler := handler.NewAuthHandler(adminRepo, cfg.JWTSecret)
	auditLogHandler := handler.NewAuditLogHandler(auditLogRepo)
	statisticsHandler := handler.NewStatisticsHandler(jamaahRepo, paketRepo)
	pengumumanHandler := handler.NewPengumumanHandler(repository.NewPengumumanRepository(db), auditLogRepo)
	manasikHandler := handler.NewManasikHandler(repository.NewManasikRepository(db), jamaahRepo, paketRepo, auditLogRepo, cfg.JWTSecret)

	// Setup Gin router
	r := gin.Default()
	// Only these proxies may set X-Forwarded-For; otherwise clients could fake
	// their IP and bypass the login / Cek Kamar rate limits.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("TRUSTED_PROXIES tidak valid: %v", err)
	}
	r.Use(middleware.SecurityHeaders())
	api := r.Group("/api")

	// Health check for Docker / load balancers.
	api.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "mongodb": "unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes (login rate limited: 30 attempts per 5 minutes per IP)
	loginLimiter := middleware.NewRateLimiter(30, 5*time.Minute)
	authHandler.RegisterPublicRoutes(api, loginLimiter.Middleware())
	// Not behind AuthRequired because <img> can't send headers; GetDocument
	// itself requires a signed link or an admin token.
	api.GET("/jamaah/:id/dokumen/:docType", jamaahHandler.GetDocument)

	// Public self-service lookup for jamaah (rate limited: 20 attempts per 5 minutes per IP)
	portalLimiter := middleware.NewRateLimiter(20, 5*time.Minute)
	handler.NewPortalHandler(jamaahRepo, paketRepo, cfg.JWTSecret).RegisterRoutes(api, portalLimiter.Middleware())
	pengumumanHandler.RegisterPublicRoutes(api)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	authHandler.RegisterProtectedRoutes(protected)
	jamaahHandler.RegisterRoutes(protected)
	paketHandler.RegisterRoutes(protected)
	auditLogHandler.RegisterRoutes(protected)
	protected.GET("/statistics", statisticsHandler.GetStatistics)
	pengumumanHandler.RegisterRoutes(protected)
	manasikHandler.RegisterRoutes(protected)
	handler.NewTagihanHandler(jamaahRepo, paketRepo).RegisterRoutes(protected)

	if cfg.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

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
	if cfg.EnableSwagger {
		log.Println("Swagger UI: http://localhost:" + cfg.AppPort + "/swagger/index.html")
	}
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
