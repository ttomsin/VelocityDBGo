package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"VelocityDBGo/internal/database"
	"VelocityDBGo/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func getCollectionOrAbort(c *gin.Context, projectId, collectionName string) *models.Collection {
	var project models.Project

	// 1. Check if authenticated via API Key (Client App)
	if apiKeyRaw, exists := c.Get("apiKey"); exists {
		apiKey := apiKeyRaw.(string)
		if err := database.DB.Where("id = ? AND api_key = ?", projectId, apiKey).First(&project).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or invalid API Key"})
			c.Abort()
			return nil
		}
	} else if userIdRaw, exists := c.Get("userId"); exists {
		// 2. Fallback to JWT Developer Authentication
		userId := uint(userIdRaw.(float64))
		if err := database.DB.Where("id = ? AND user_id = ?", projectId, userId).First(&project).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
			c.Abort()
			return nil
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
		return nil
	}

	var collection models.Collection
	if err := database.DB.Where("project_id = ? AND name = ?", project.ID, collectionName).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		c.Abort()
		return nil
	}
	return &collection
}

// InsertDocument godoc
// @Summary      Insert JSON document
// @Description  Insert an arbitrary JSON payload into a collection
// @Tags         Data
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        request body map[string]interface{} true "JSON Data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName} [post]
func InsertDocument(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process JSON"})
		return
	}

	doc := models.Document{
		CollectionID: collection.ID,
		Data:         datatypes.JSON(jsonBytes),
	}

	if err := database.DB.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document inserted", "id": doc.ID, "data": jsonData})
}

// buildJSONBPath converts dot notation like "address.city" into PostgreSQL JSONB syntax "data->'address'->>'city'".
// This allows students to query deeply nested JSON structures intuitively.
func buildJSONBPath(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		// Single key: just extract the text value directly from the data column
		return fmt.Sprintf("data->>'%s'", parts[0])
	}

	path := "data"
	for i, part := range parts {
		if i == len(parts)-1 {
			// The final node uses ->> to return the value as text so we can compare it (e.g. =, >, <)
			path += fmt.Sprintf("->>'%s'", part)
		} else {
			// Intermediate nodes use -> to return the JSONB object itself to keep traversing
			path += fmt.Sprintf("->'%s'", part)
		}
	}
	return path
}

// GetDocuments godoc
// @Summary      Get all documents
// @Description  Retrieve all documents from a collection. Supports dynamic filtering via query parameters (e.g. ?name=John, ?address.city=London, ?age=gte:18), sorting (e.g. ?sort=address.zip:desc), and pagination (?limit=10&offset=0).
// @Tags         Data
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        limit query int false "Pagination limit"
// @Param        offset query int false "Pagination offset"
// @Param        sort query string false "Sort field and direction (e.g., age:desc, address.city:asc)"
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName} [get]
func GetDocuments(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	query := database.DB.Where("collection_id = ?", collection.ID)

	// Parse query parameters for dynamic JSONB filtering
	// Example URL: /data/users?address.city=London&age=gt:18
	for key, values := range c.Request.URL.Query() {
		// Skip reserved query parameters used for pagination and sorting
		if key == "limit" || key == "offset" || key == "sort" {
			continue
		}

		if len(values) > 0 {
			value := values[0]

			// Transform "address.city" into "data->'address'->>'city'"
			jsonbPath := buildJSONBPath(key)

			// Simple parsing for SQL operators (gt, lt, like, neq, etc.)
			// We cast the JSONB text value to NUMERIC for mathematical comparisons (<, >, <=, >=)
			if strings.HasPrefix(value, "eq:") {
				query = query.Where(fmt.Sprintf("%s = ?", jsonbPath), value[3:])
			} else if strings.HasPrefix(value, "neq:") {
				query = query.Where(fmt.Sprintf("%s != ?", jsonbPath), value[4:])
			} else if strings.HasPrefix(value, "gt:") {
				query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) > ?", jsonbPath), value[3:])
			} else if strings.HasPrefix(value, "gte:") {
				query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) >= ?", jsonbPath), value[4:])
			} else if strings.HasPrefix(value, "lt:") {
				query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) < ?", jsonbPath), value[3:])
			} else if strings.HasPrefix(value, "lte:") {
				query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) <= ?", jsonbPath), value[4:])
			} else if strings.HasPrefix(value, "like:") {
				// ILIKE provides case-insensitive fuzzy matching
				query = query.Where(fmt.Sprintf("%s ILIKE ?", jsonbPath), fmt.Sprintf("%%%s%%", value[5:]))
			} else {
				// Default behavior is equality checking if no operator prefix is provided
				query = query.Where(fmt.Sprintf("%s = ?", jsonbPath), value)
			}
		}
	}

	// Add Pagination (Limit & Offset)
	limit := c.Query("limit")
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			query = query.Limit(l)
		}
	}

	offset := c.Query("offset")
	if offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			query = query.Offset(o)
		}
	}

	// Add Sorting (e.g., ?sort=address.city:desc)
	sort := c.Query("sort")
	if sort != "" {
		parts := strings.Split(sort, ":")
		fieldPath := buildJSONBPath(parts[0])

		if len(parts) == 2 {
			order := parts[1]
			if strings.ToLower(order) == "desc" {
				query = query.Order(fmt.Sprintf("%s DESC", fieldPath))
			} else {
				query = query.Order(fmt.Sprintf("%s ASC", fieldPath))
			}
		} else {
			// Default to ascending order if no direction is specified
			query = query.Order(fmt.Sprintf("%s ASC", fieldPath))
		}
	}

	var documents []models.Document
	query.Find(&documents)

	var response []map[string]interface{}
	for _, doc := range documents {
		var data map[string]interface{}
		json.Unmarshal(doc.Data, &data)
		data["_id"] = doc.ID // Inject document ID into the returned JSON
		response = append(response, data)
	}

	if response == nil {
		response = make([]map[string]interface{}, 0)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDocument godoc
// @Summary      Delete a document
// @Description  Delete a document from a collection by ID
// @Tags         Data
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        docId path int true "Document ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName}/{docId} [delete]
func DeleteDocument(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")
	docIdStr := c.Param("docId")

	docId, err := strconv.ParseUint(docIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	if err := database.DB.Where("id = ? AND collection_id = ?", docId, collection.ID).Delete(&models.Document{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

type QueryRequest struct {
	Filter map[string]interface{} `json:"filter"` // Example: {"address.city": "London", "age": "gt:18"}
	Sort   string                 `json:"sort" example:"address.city:desc"`
	Limit  int                    `json:"limit" example:"10"`
	Offset int                    `json:"offset" example:"0"`
}

// QueryDocuments godoc
// @Summary      Query documents via JSON
// @Description  Query documents using a structured JSON POST payload. Useful when URL query parameters become too complex.
// @Tags         Data
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        request body QueryRequest true "Query Parameters"
// @Success      200  {array}   map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName}/query [post]
func QueryDocuments(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query payload"})
		return
	}

	query := database.DB.Where("collection_id = ?", collection.ID)

	// Apply JSONB filters
	for key, val := range req.Filter {
		jsonbPath := buildJSONBPath(key)
		value := fmt.Sprintf("%v", val)

		if strings.HasPrefix(value, "eq:") {
			query = query.Where(fmt.Sprintf("%s = ?", jsonbPath), value[3:])
		} else if strings.HasPrefix(value, "neq:") {
			query = query.Where(fmt.Sprintf("%s != ?", jsonbPath), value[4:])
		} else if strings.HasPrefix(value, "gt:") {
			query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) > ?", jsonbPath), value[3:])
		} else if strings.HasPrefix(value, "gte:") {
			query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) >= ?", jsonbPath), value[4:])
		} else if strings.HasPrefix(value, "lt:") {
			query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) < ?", jsonbPath), value[3:])
		} else if strings.HasPrefix(value, "lte:") {
			query = query.Where(fmt.Sprintf("CAST(%s AS NUMERIC) <= ?", jsonbPath), value[4:])
		} else if strings.HasPrefix(value, "like:") {
			query = query.Where(fmt.Sprintf("%s ILIKE ?", jsonbPath), fmt.Sprintf("%%%s%%", value[5:]))
		} else {
			query = query.Where(fmt.Sprintf("%s = ?", jsonbPath), value)
		}
	}

	// Pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Sorting
	if req.Sort != "" {
		parts := strings.Split(req.Sort, ":")
		fieldPath := buildJSONBPath(parts[0])
		if len(parts) == 2 {
			order := parts[1]
			if strings.ToLower(order) == "desc" {
				query = query.Order(fmt.Sprintf("%s DESC", fieldPath))
			} else {
				query = query.Order(fmt.Sprintf("%s ASC", fieldPath))
			}
		} else {
			query = query.Order(fmt.Sprintf("%s ASC", fieldPath))
		}
	}

	var documents []models.Document
	query.Find(&documents)

	var response []map[string]interface{}
	for _, doc := range documents {
		var data map[string]interface{}
		json.Unmarshal(doc.Data, &data)
		data["_id"] = doc.ID // Inject document ID
		response = append(response, data)
	}

	if response == nil {
		response = make([]map[string]interface{}, 0)
	}

	c.JSON(http.StatusOK, response)
}

// GetDocument godoc
// @Summary      Get a single document
// @Description  Retrieve a specific document from a collection by ID
// @Tags         Data
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        docId path int true "Document ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName}/{docId} [get]
func GetDocument(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")
	docIdStr := c.Param("docId")

	docId, err := strconv.ParseUint(docIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	var doc models.Document
	if err := database.DB.Where("id = ? AND collection_id = ?", docId, collection.ID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	var data map[string]interface{}
	json.Unmarshal(doc.Data, &data)
	data["_id"] = doc.ID // Inject document ID into the returned JSON

	c.JSON(http.StatusOK, data)
}

// UpdateDocument godoc
// @Summary      Update a document
// @Description  Fully replace or update the JSON payload of an existing document
// @Tags         Data
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Param        projectId path int true "Project ID"
// @Param        collectionName path string true "Collection Name"
// @Param        docId path int true "Document ID"
// @Param        request body map[string]interface{} true "Updated JSON Data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/projects/{projectId}/data/{collectionName}/{docId} [put]
func UpdateDocument(c *gin.Context) {
	projectId := c.Param("projectId")
	collectionName := c.Param("collectionName")
	docIdStr := c.Param("docId")

	docId, err := strconv.ParseUint(docIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	collection := getCollectionOrAbort(c, projectId, collectionName)
	if collection == nil {
		return
	}

	// Verify document exists
	var doc models.Document
	if err := database.DB.Where("id = ? AND collection_id = ?", docId, collection.ID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Parse new JSON payload
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process JSON"})
		return
	}

	// Update document
	doc.Data = datatypes.JSON(jsonBytes)
	if err := database.DB.Save(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document updated", "id": doc.ID, "data": jsonData})
}
