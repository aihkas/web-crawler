package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"web-crawler/internal/models"
	"strings"

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

func GetAllAnalyses(db *sql.DB) ([]models.Analysis, error) {
	rows, err := db.Query("SELECT id, url, status, error_msg, page_title, html_version, heading_counts, internal_link_count, external_link_count, inaccessible_links, has_login_form, created_at, updated_at FROM analysis_results ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query analyses: %w", err)
	}
	defer rows.Close()

	var analyses []models.Analysis
	for rows.Next() {
		var analysis models.Analysis
		var headingsJSON, inaccessibleLinksJSON []byte // Raw JSON from DB

		err := rows.Scan(
			&analysis.ID, &analysis.URL, &analysis.Status, &analysis.ErrorMsg,
			&analysis.PageTitle, &analysis.HTMLVersion, &headingsJSON,
			&analysis.InternalLinkCount, &analysis.ExternalLinkCount,
			&inaccessibleLinksJSON, &analysis.HasLoginForm,
			&analysis.CreatedAt, &analysis.UpdatedAt,
		)
		if err != nil {
			log.Printf("Warning: Failed to scan row: %v", err)
			continue
		}
		
		if headingsJSON != nil {
			json.Unmarshal(headingsJSON, &analysis.HeadingCounts)
		}
		if inaccessibleLinksJSON != nil {
			json.Unmarshal(inaccessibleLinksJSON, &analysis.InaccessibleLinks)
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// GetAnalysisByID retrieves a single analysis record by its ID.
func GetAnalysisByID(db *sql.DB, id int64) (*models.Analysis, error) {
	row := db.QueryRow("SELECT id, url, status, error_msg, page_title, html_version, heading_counts, internal_link_count, external_link_count, inaccessible_links, has_login_form, created_at, updated_at FROM analysis_results WHERE id = ?", id)

	var analysis models.Analysis
	var headingsJSON, inaccessibleLinksJSON []byte

	err := row.Scan(
		&analysis.ID, &analysis.URL, &analysis.Status, &analysis.ErrorMsg,
		&analysis.PageTitle, &analysis.HTMLVersion, &headingsJSON,
		&analysis.InternalLinkCount, &analysis.ExternalLinkCount,
		&inaccessibleLinksJSON, &analysis.HasLoginForm,
		&analysis.CreatedAt, &analysis.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analysis with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if headingsJSON != nil {
		json.Unmarshal(headingsJSON, &analysis.HeadingCounts)
	}
	if inaccessibleLinksJSON != nil {
		json.Unmarshal(inaccessibleLinksJSON, &analysis.InaccessibleLinks)
	}

	return &analysis, nil
}

// BulkDeleteAnalyses deletes records from analysis_results by their IDs.
func BulkDeleteAnalyses(db *sql.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// Prepare the query with the correct number of placeholders (?) to prevent SQL injection.
	query := "DELETE FROM analysis_results WHERE id IN (?" + strings.Repeat(",?", len(ids)-1) + ")"
	
	// Convert []int64 to []interface{} for the Exec function.
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}

	return nil
}