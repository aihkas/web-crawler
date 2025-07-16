CREATE TABLE IF NOT EXISTS analysis_results (
    id INT AUTO_INCREMENT PRIMARY KEY,
    url TEXT NOT NULL,
    status ENUM('queued', 'running', 'done', 'error') NOT NULL DEFAULT 'queued',
    error_msg TEXT,
    page_title TEXT,
    html_version VARCHAR(50),
    heading_counts JSON,
    internal_link_count INT,
    external_link_count INT,
    inaccessible_links JSON,
    has_login_form BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
