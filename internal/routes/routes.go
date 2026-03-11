package routes

import (
	_ "VelocityDBGo/docs" // This import is required to initialize the swagger docs
	"VelocityDBGo/internal/handlers"
	"VelocityDBGo/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Frontend Dashboard routes
	r.Static("/assets", "./public/assets")
	r.StaticFile("/", "./public/index.html")
	r.StaticFile("/auth", "./public/auth.html")
	r.StaticFile("/dashboard", "./public/dashboard.html")
	r.StaticFile("/docs", "./public/docs.html")

	// Public routes
	r.POST("/api/auth/signup", handlers.Signup)
	r.POST("/api/auth/login", handlers.Login)

	// Protected platform routes (for developers/students)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Projects
		protected.POST("/projects", handlers.CreateProject)
		protected.GET("/projects", handlers.GetProjects)

		// Collections within a project
		protected.POST("/projects/:projectId/collections", handlers.CreateCollection)
		protected.GET("/projects/:projectId/collections", handlers.GetCollections)
	}

	// Data API routes (can be accessed via Developer JWT OR Project API Key)
	dataApi := r.Group("/api")
	dataApi.Use(middleware.AuthOrAPIKeyMiddleware())
	{
		dataApi.POST("/projects/:projectId/data/:collectionName", handlers.InsertDocument)
		dataApi.GET("/projects/:projectId/data/:collectionName", handlers.GetDocuments)
		dataApi.GET("/projects/:projectId/data/:collectionName/:docId", handlers.GetDocument)
		dataApi.PUT("/projects/:projectId/data/:collectionName/:docId", handlers.UpdateDocument)
		dataApi.POST("/projects/:projectId/data/:collectionName/query", handlers.QueryDocuments)
		dataApi.DELETE("/projects/:projectId/data/:collectionName/:docId", handlers.DeleteDocument)
	}
}
