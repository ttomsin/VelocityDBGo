package handlers

import (
	"net/http"
	"strconv"

	"VelocityDBGo/internal/database"
	"VelocityDBGo/internal/models"
	"github.com/gin-gonic/gin"
)

type CreateCollectionRequest struct {
	Name string `json:"name" binding:"required" example:"restaurants"`
}

// CreateCollection godoc
// @Summary      Create a new collection
// @Description  Create a virtual table within a specific project
// @Tags         Collections
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId path int true "Project ID"
// @Param        request body CreateCollectionRequest true "Collection parameters"
// @Success      201  {object}  models.Collection
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/collections [post]
func CreateCollection(c *gin.Context) {
	projectIdParam := c.Param("projectId")
	projectId, err := strconv.ParseUint(projectIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the project belongs to the user
	userIdRaw, _ := c.Get("userId")
	userId := uint(userIdRaw.(float64))

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", projectId, userId).First(&project).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	collection := models.Collection{
		ProjectID: uint(projectId),
		Name:      req.Name,
	}

	if err := database.DB.Where("project_id = ? AND name = ?", projectId, req.Name).First(&models.Collection{}).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection name already exists in this project"})
		return
	}

	if err := database.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, collection)
}

// GetCollections godoc
// @Summary      Get all collections
// @Description  Retrieve all collections for a specific project
// @Tags         Collections
// @Produce      json
// @Security     BearerAuth
// @Param        projectId path int true "Project ID"
// @Success      200  {array}   models.Collection
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/collections [get]
func GetCollections(c *gin.Context) {
	projectIdParam := c.Param("projectId")
	projectId, err := strconv.ParseUint(projectIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project ownership
	userIdRaw, _ := c.Get("userId")
	userId := uint(userIdRaw.(float64))

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", projectId, userId).First(&project).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	var collections []models.Collection
	database.DB.Where("project_id = ?", projectId).Find(&collections)

	c.JSON(http.StatusOK, collections)
}
