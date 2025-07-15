package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv" 

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"web-crawler/internal/crawler"
	"web-crawler/internal/database"
)

type Server struct {
	db *sql.DB
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS environment variables")
	}
}

// AuthMiddleware provides token-based authentication for API routes.
func AuthMiddleware() gin.HandlerFunc {
	requiredToken := os.Getenv("API_TOKEN")
	if requiredToken == "" {
		log.Fatal("FATAL: API_TOKEN environment variable not set.")
	}
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in 'Bearer {token}' format"})
			return
		}
		if parts[1] != requiredToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
			return
		}
		c.Next()
	}
}

func main() {
	loadEnv()
	db := database.InitDB()
	defer db.Close()

	server := &Server{db: db}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api/v1")
	api.Use(AuthMiddleware())
	{
		api.POST("/analyze", server.handleAnalyzeRequest)
		api.GET("/results", server.handleGetResults)
		api.GET("/results/:id", server.handleGetResultByID)
		api.DELETE("/results", server.handleBulkDelete)
	}
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	log.Println("Server starting on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type AnalyzeRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// handleAnalyzeRequest receives a URL, creates a job record,
// and starts the analysis in the background.
func (s *Server) handleAnalyzeRequest(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	//Create a record in the database with 'queued' status.
	id, err := database.CreateAnalysis(s.db, req.URL)
	if err != nil {
		log.Printf("ERROR: Failed to create analysis record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create analysis job"})
		return
	}

	//Start the analysis in a new goroutine so the API can respond immediately.
	go s.runAnalysis(id, req.URL)

	//Respond with 202 Accepted to indicate the request has been received.
	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Analysis request accepted and is being processed.",
		"analysis_id": id,
	})
}

// runAnalysis is the background job that performs the web page crawling.
func (s *Server) runAnalysis(id int64, url string) {
	log.Printf("INFO: Starting analysis for job ID %d, URL: %s", id, url)

	// Update status to 'running'
	if err := database.UpdateAnalysisStatus(s.db, id, "running", ""); err != nil {
		log.Printf("ERROR: Failed to update status to 'running' for job ID %d: %v", id, err)
		return
	}

	// Perform the analysis
	result, err := crawler.AnalyzeURL(url)
	if err != nil {
		log.Printf("ERROR: Analysis failed for job ID %d: %v", id, err)
		if updateErr := database.UpdateAnalysisStatus(s.db, id, "error", err.Error()); updateErr != nil {
			log.Printf("ERROR: Failed to update status to 'error' for job ID %d: %v", id, updateErr)
		}
		return
	}

	// Save the successful results to the database and set status to 'done'
	if err := database.SaveAnalysisResult(s.db, id, result); err != nil {
		log.Printf("ERROR: Failed to save analysis results for job ID %d: %v", id, err)
		if updateErr := database.UpdateAnalysisStatus(s.db, id, "error", "Failed to save results"); updateErr != nil {
			log.Printf("ERROR: Failed to update status to 'error' for job ID %d: %v", id, updateErr)
		}
		return
	}

	log.Printf("INFO: Successfully completed analysis for job ID %d", id)
}

func (s *Server) handleGetResults(c *gin.Context) {
	results, err := database.GetAllAnalyses(s.db)
	if err != nil {
		log.Printf("ERROR: Failed to get analysis results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve results"})
		return
	}
	c.JSON(http.StatusOK, results)
}

// handleGetResultByID fetches a single analysis by its ID.
func (s *Server) handleGetResultByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	result, err := database.GetAnalysisByID(s.db, id)
	if err != nil {
		// Differentiate between not found and other errors
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("ERROR: Failed to get analysis by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve result"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

type BulkDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// handleBulkDelete handles the deletion of multiple analysis records.
func (s *Server) handleBulkDelete(c *gin.Context) {
	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided for deletion"})
		return
	}

	err := database.BulkDeleteAnalyses(s.db, req.IDs)
	if err != nil {
		log.Printf("ERROR: Failed to bulk delete analyses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted records"})
}
