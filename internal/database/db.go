package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"web-crawler/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to create database handle: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established successfully.")
	return db
}

// CreateAnalysis inserts a new analysis request into the database with 'queued' status.
func CreateAnalysis(db *sql.DB, url string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO analysis_results(url, status) VALUES(?, 'queued')")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(url)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// UpdateAnalysisStatus updates the status of an analysis record.
func UpdateAnalysisStatus(db *sql.DB, id int64, status string, errorMsg string) error {
	stmt, err := db.Prepare("UPDATE analysis_results SET status = ?, error_msg = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, errorMsg, id)
	return err
}

// SaveAnalysisResult saves the full results of a successful analysis.
func SaveAnalysisResult(db *sql.DB, id int64, result *models.Analysis) error {
	headingsJSON, err := json.Marshal(result.HeadingCounts)
	if err != nil {
		return fmt.Errorf("failed to marshal heading counts: %w", err)
	}
	
	inaccessibleLinksJSON, err := json.Marshal(result.InaccessibleLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal inaccessible links: %w", err)
	}

	query := `
		UPDATE analysis_results 
		SET
			status = 'done',
			page_title = ?,
			html_version = ?,
			heading_counts = ?,
			internal_link_count = ?,
			external_link_count = ?,
			inaccessible_links = ?,
			has_login_form = ?
		WHERE id = ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare result save statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		result.PageTitle,
		result.HTMLVersion,
		headingsJSON,
		result.InternalLinkCount,
		result.ExternalLinkCount,
		inaccessibleLinksJSON,
		result.HasLoginForm,
		id,
	)

	return err
}
