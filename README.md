# Web Page Analyzer

Web Page Analyzer is a full-stack web application designed to crawl a given website URL and provide a detailed analysis of its on-page SEO and technical elements. It features a Go-based backend for high-performance crawling and a modern React/TypeScript frontend for a responsive and interactive user experience.

The entire application stack is containerized with Docker, allowing for a simple, one-command setup.

## Features

* **URL Analysis:** Submit any public URL to initiate a detailed analysis.
* **Asynchronous Processing:** Uses goroutines on the backend to process analysis jobs in the background without blocking the UI.
* **Comprehensive Data Extraction:**
    * HTML version detection.
    * Page title extraction.
    * Counts of heading tags (H1, H2, H3, etc.).
    * Ratio of internal vs. external links.
    * Detection of a login form (`<input type="password">`).
    * **Inaccessible Link Checking:** Concurrently checks all found links and reports any that return a 4xx or 5xx status code.
* **Interactive Dashboard:**
    * Real-time status updates for analysis jobs (`queued`, `running`, `done`, `error`).
    * A sortable, filterable table to view all results.
    * Bulk actions, including the ability to delete multiple results at once.
* **Detailed View:**
    * A dedicated page for each analysis result.
    * A donut chart visualizing the distribution of internal vs. external links.
    * A clear list of all broken links with their corresponding HTTP status codes.
* **Secure API:** The backend API is protected by token-based authentication.


## Tech Stack

| Category | Technology |
| :-- | :-- |
| **Backend** | Go (Golang) with the Gin web framework |
| **Frontend** | React, TypeScript, TanStack Table, Recharts |
| **Database** | MySQL 8.0 |
| **DevOps** | Docker \& Docker Compose |
| **Testing** | React Testing Library \& Jest |

## Prerequisites

To run this project, you will need:

* Docker
* Docker Compose

For manual local development, you will also need:

* Go (version 1.22 or later)
* Node.js (version 20.x or later)
* A running instance of MySQL


## Getting Started

### Using Docker (Recommended Method)

This is the simplest way to get the entire application stack (backend, frontend, and database) running.

1. **Clone the repository:**

```bash
git clone <your-repository-url>
cd <repository-folder>
```

2. **Configure Environment Variables:**
Copy the `.env.example` file to a new file named `.env` and fill in the required values. The `DB_HOST` must be set to `db` to connect to the Docker container.

```bash
cp .env.example .env
```

Your `.env` file should look like this:

```
DB_HOST="db"
DB_PORT="3306"
DB_USER="your_user"
DB_PASSWORD="your_password"
DB_NAME="web_crawler_db"
DB_ROOT_PASSWORD="your_strong_root_password"
API_TOKEN="your-secret-api-token"
```

3. **Build and Run with Docker Compose:**
From the root directory, run the following command:

```bash
docker-compose up --build
```

This will build the images, start all the containers, and automatically set up the database table.
4. **Access the Application:**
Open your browser and navigate to **`http://localhost:3000`**.

### Manual Setup (for Development)

Follow these steps if you want to run the backend and frontend servers independently.

#### Backend Server:

1. Navigate to the project's root directory.
2. Install Go dependencies:

```bash
go mod tidy
```

3. Ensure your `.env` file is configured to point to your local MySQL instance (e.g., `DB_HOST="127.0.0.1"`).
4. Run the server:

```bash
go run ./cmd/server
```

The backend will be running on `http://localhost:8080`.

#### Frontend Server:

1. Navigate to the `frontend` directory.
2. Install Node.js dependencies:

```bash
npm install
```

3. Create a `.env` file in the `frontend` directory with your API token.

```
REACT_APP_API_BASE_URL=http://localhost:8080/api/v1
REACT_APP_API_TOKEN=your-secret-api-token
```

4. Run the development server:

```bash
npm start
```

The frontend will be running on `http://localhost:3000`.

## API Overview

All endpoints are prefixed with `/api/v1` and require an `Authorization: Bearer <your_token>` header.


| Method | Endpoint | Description |
| :-- | :-- | :-- |
| `POST` | `/analyze` | Submits a new URL for analysis. |
| `GET` | `/results` | Retrieves all analysis results. |
| `GET` | `/results/:id` | Retrieves a single analysis result by its ID. |
| `DELETE` | `/results` | Deletes one or more analysis results by ID. |

## To Do's And Areas For Improvment

[] Write  tests.
[] Implement progress bar.
[] Use an Nginx reverse proxy to route API calls.
[] Server-Side Pagination and Filtering.
[] Use of Database Transactions.
[] Advanced State Management & Data Fetching.
[] Scoped and Maintainable CSS.
