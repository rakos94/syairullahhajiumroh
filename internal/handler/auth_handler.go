package handler

import (
	"net/http"
	"time"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo      *repository.AdminRepository
	jwtSecret string
}

func NewAuthHandler(repo *repository.AdminRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{repo: repo, jwtSecret: jwtSecret}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.repo.FindByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "username atau password salah"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username atau password salah"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      admin.ID.Hex(),
		"username": admin.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *AuthHandler) Me(c *gin.Context) {
	admin, ok := h.getCurrentAdmin(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, admin)
}

func (h *AuthHandler) getCurrentAdmin(c *gin.Context) (*model.Admin, bool) {
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return nil, false
	}
	adminID, err := primitive.ObjectIDFromHex(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid"})
		return nil, false
	}
	admin, err := h.repo.FindByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin tidak ditemukan"})
		return nil, false
	}
	return admin, true
}

func (h *AuthHandler) ListAdmins(c *gin.Context) {
	admins, err := h.repo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admins)
}

type adminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	current, ok := h.getCurrentAdmin(c)
	if !ok {
		return
	}
	if current.Role != "super" {
		c.JSON(http.StatusForbidden, gin.H{"error": "hanya super admin yang dapat menambah admin"})
		return
	}

	var req adminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal hash password"})
		return
	}

	role := req.Role
	if role == "" {
		role = "admin"
	}

	admin := &model.Admin{
		Username: req.Username,
		Password: string(hashed),
		Role:     role,
	}
	if err := h.repo.Create(c.Request.Context(), admin); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, admin)
}

type adminUpdateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *AuthHandler) UpdateAdmin(c *gin.Context) {
	current, ok := h.getCurrentAdmin(c)
	if !ok {
		return
	}
	if current.Role != "super" {
		c.JSON(http.StatusForbidden, gin.H{"error": "hanya super admin yang dapat mengubah admin"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req adminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Username != "" {
		update["username"] = req.Username
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal hash password"})
			return
		}
		update["password"] = string(hashed)
	}
	if req.Role != "" {
		update["role"] = req.Role
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak ada data yang diubah"})
		return
	}

	if err := h.repo.Update(c.Request.Context(), id, update); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username sudah digunakan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	admin, _ := h.repo.FindByID(c.Request.Context(), id)
	c.JSON(http.StatusOK, admin)
}

func (h *AuthHandler) DeleteAdmin(c *gin.Context) {
	current, ok := h.getCurrentAdmin(c)
	if !ok {
		return
	}
	if current.Role != "super" {
		c.JSON(http.StatusForbidden, gin.H{"error": "hanya super admin yang dapat menghapus admin"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	target, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin tidak ditemukan"})
		return
	}
	if target.Role == "super" {
		c.JSON(http.StatusForbidden, gin.H{"error": "tidak dapat menghapus super admin"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "admin berhasil dihapus"})
}

func (h *AuthHandler) RegisterPublicRoutes(r *gin.RouterGroup, loginMiddlewares ...gin.HandlerFunc) {
	auth := r.Group("/auth")
	{
		handlers := append(loginMiddlewares, h.Login)
		auth.POST("/login", handlers...)
	}
}

func (h *AuthHandler) RegisterProtectedRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/me", h.Me)
	}
	admin := r.Group("/admin")
	{
		admin.GET("", h.ListAdmins)
		admin.POST("", h.CreateAdmin)
		admin.PUT("/:id", h.UpdateAdmin)
		admin.DELETE("/:id", h.DeleteAdmin)
	}
}
