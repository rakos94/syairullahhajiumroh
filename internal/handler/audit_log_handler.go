package handler

import (
	"net/http"
	"strconv"

	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLogHandler struct {
	repo *repository.AuditLogRepository
}

func NewAuditLogHandler(repo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

func (h *AuditLogHandler) FindAll(c *gin.Context) {
	entityType := c.Query("entity_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.repo.FindAll(c.Request.Context(), entityType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": repository.TotalPages(total, limit),
	})
}

func (h *AuditLogHandler) FindByEntity(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID, err := primitive.ObjectIDFromHex(c.Param("entityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.repo.FindByEntity(c.Request.Context(), entityType, entityID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": repository.TotalPages(total, limit),
	})
}

func (h *AuditLogHandler) RegisterRoutes(r *gin.RouterGroup) {
	audit := r.Group("/audit-logs")
	{
		audit.GET("", h.FindAll)
		audit.GET("/:entityType/:entityId", h.FindByEntity)
	}
}
