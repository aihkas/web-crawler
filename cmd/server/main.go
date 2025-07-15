package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

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

	// Apply the authentication middleware to the entire /api/v1 route group.
	api := router.Group("/api/v1")
	api.Use(AuthMiddleware())
	{
		api.POST("/analyze", server.handleAnalyzeRequest)
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
