package handlers

import (
	"net/http"

	"VelocityDBGo/internal/database"
	"VelocityDBGo/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required" example:"Foodie"`
	Description string `json:"description" example:"My first food delivery app"`
}

// CreateProject godoc
// @Summary      Create a new project
// @Description  Create a logical container for an application
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateProjectRequest true "Project parameters"
// @Success      201  {object}  models.Project
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects [post]
func CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdRaw, _ := c.Get("userId")
	userId := uint(userIdRaw.(float64))

	project := models.Project{
		UserID:      userId,
		Name:        req.Name,
		Description: req.Description,
		APIKey:      uuid.New().String(),
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjects godoc
// @Summary      Get all projects
// @Description  Retrieve all projects belonging to the authenticated user
// @Tags         Projects
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Project
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/projects [get]
func GetProjects(c *gin.Context) {
	userIdRaw, _ := c.Get("userId")
	userId := uint(userIdRaw.(float64))

	var projects []models.Project
	database.DB.Where("user_id = ?", userId).Find(&projects)

	c.JSON(http.StatusOK, projects)
}

type UpdateProjectStatusRequest struct {
	IsActive bool `json:"isActive"`
}

// UpdateProjectStatus godoc
// @Summary      Enable/Disable a project
// @Description  Toggle whether a project is active or disabled (soft delete). Disabled projects cannot accept Data API queries.
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        projectId path int true "Project ID"
// @Param        request body UpdateProjectStatusRequest true "Status parameter"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/status [patch]
func UpdateProjectStatus(c *gin.Context) {
	projectId := c.Param("projectId")

	var req UpdateProjectStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userIdRaw, _ := c.Get("userId")
	userId := uint(userIdRaw.(float64))

	var project models.Project
	if err := database.DB.Where("id = ? AND user_id = ?", projectId, userId).First(&project).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	project.IsActive = req.IsActive
	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project status"})
		return
	}

	status := "disabled"
	if project.IsActive {
		status = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project " + status, "isActive": project.IsActive})
}
