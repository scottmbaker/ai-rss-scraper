package storage

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Article struct {
	GUID          string
	Title         string
	Link          string
	Description   string
	Content       string
	PublishedDate time.Time
	Score         string
	Analysis      string
	FeedURL       string
	Model         string
	Reported      bool
}

type DB struct {
	conn *sql.DB
}

// closeRowsBOF closes the rows and bails on failure.
func closeRowsBOF(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Fatalf("Error closing rows: %v", err)
	}
}

// NewDatabase initializes the SQLite database.
func NewDatabase(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS articles (
		guid TEXT PRIMARY KEY,
		title TEXT,
		link TEXT,
		description TEXT,
		content TEXT,
		published_date DATETIME,
		score TEXT,
		analysis TEXT,
		feed_url TEXT,
		model TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		reported BOOLEAN DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	// Migration: Add reported column if it doesn't exist
	_, err = db.Exec("ALTER TABLE articles ADD COLUMN reported BOOLEAN DEFAULT 0")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}

	return &DB{conn: db}, nil
}

// ArticleExists checks if an article with the given GUID already exists.
func (d *DB) ArticleExists(guid string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE guid=?)"
	err := d.conn.QueryRow(query, guid).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// SaveArticle saves a new article to the database.
func (d *DB) SaveArticle(article Article) error {
	query := `INSERT INTO articles (guid, title, link, description, content, published_date, score, analysis, feed_url, model) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.conn.Exec(query, article.GUID, article.Title, article.Link, article.Description, article.Content, article.PublishedDate, article.Score, article.Analysis, article.FeedURL, article.Model)
	return err
}

// GetArticlesToScore retrieves articles to score.
// If refreshPattern is empty, it returns unscored articles.
// If refreshPattern is provided, it returns articles matching the title pattern (glob) OR unscored articles.
func (d *DB) GetArticlesToScore(refreshPattern string) ([]Article, error) {
	baseQuery := `SELECT guid, title, link, description, content, published_date, score, analysis, feed_url, model 
              FROM articles WHERE score IS NULL OR score = '' OR score = 'N/A'`

	var rows *sql.Rows
	var err error

	if refreshPattern != "" {
		// Convert glob to SQL LIKE
		// * -> %
		// ? -> _
		likePattern := strings.ReplaceAll(refreshPattern, "*", "%")
		likePattern = strings.ReplaceAll(likePattern, "?", "_")

		query := `SELECT guid, title, link, description, content, published_date, score, analysis, feed_url, model 
              FROM articles WHERE (score IS NULL OR score = '' OR score = 'N/A') OR title LIKE ?`
		rows, err = d.conn.Query(query, likePattern)
	} else {
		rows, err = d.conn.Query(baseQuery)
	}

	if err != nil {
		return nil, err
	}
	defer closeRowsBOF(rows)

	var articles []Article
	for rows.Next() {
		var art Article
		err := rows.Scan(&art.GUID, &art.Title, &art.Link, &art.Description, &art.Content, &art.PublishedDate, &art.Score, &art.Analysis, &art.FeedURL, &art.Model)
		if err != nil {
			return nil, err
		}
		articles = append(articles, art)
	}
	return articles, nil
}

// UpdateArticleScore updates the score and analysis for a given article GUID.
func (d *DB) UpdateArticleScore(guid, score, analysis, model string) error {
	query := `UPDATE articles SET score = ?, analysis = ?, model = ? WHERE guid = ?`
	_, err := d.conn.Exec(query, score, analysis, model, guid)
	return err
}

// ListArticles retrieves the most recent articles, up to the specified limit.
func (d *DB) ListArticles(limit int) ([]Article, error) {
	return d.ListArticlesFiltered(limit, false)
}

// ListArticlesFiltered retrieves the most recent articles, optionally filtering by reported status.
func (d *DB) ListArticlesFiltered(limit int, reportedOnly bool) ([]Article, error) {
	var query string
	if reportedOnly {
		query = `SELECT guid, title, link, description, COALESCE(content, ''), published_date, score, analysis, COALESCE(feed_url, ''), COALESCE(model, ''), reported
              FROM articles WHERE reported = 1 ORDER BY published_date DESC LIMIT ?`
	} else {
		query = `SELECT guid, title, link, description, COALESCE(content, ''), published_date, score, analysis, COALESCE(feed_url, ''), COALESCE(model, ''), reported
              FROM articles ORDER BY published_date DESC LIMIT ?`
	}

	rows, err := d.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer closeRowsBOF(rows)

	var articles []Article
	for rows.Next() {
		var art Article
		err := rows.Scan(&art.GUID, &art.Title, &art.Link, &art.Description, &art.Content, &art.PublishedDate, &art.Score, &art.Analysis, &art.FeedURL, &art.Model, &art.Reported)
		if err != nil {
			return nil, err
		}
		articles = append(articles, art)
	}
	return articles, nil
}

// GetArticlesAfter retrieves articles published after the specified time.
func (d *DB) GetArticlesAfter(since time.Time) ([]Article, error) {
	query := `SELECT guid, title, link, description, content, published_date, score, analysis, feed_url, model, reported 
              FROM articles WHERE published_date >= ? ORDER BY published_date DESC`
	rows, err := d.conn.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer closeRowsBOF(rows)

	var articles []Article
	for rows.Next() {
		var art Article

		err := rows.Scan(&art.GUID, &art.Title, &art.Link, &art.Description, &art.Content, &art.PublishedDate, &art.Score, &art.Analysis, &art.FeedURL, &art.Model, &art.Reported)
		if err != nil {
			return nil, err
		}
		articles = append(articles, art)
	}
	return articles, nil
}

// GetUnreportedArticlesAfter retrieves unreported articles published after the specified time.
func (d *DB) GetUnreportedArticlesAfter(since time.Time) ([]Article, error) {
	query := `SELECT guid, title, link, description, content, published_date, score, analysis, feed_url, model, reported 
              FROM articles WHERE published_date >= ? AND (reported = 0 OR reported IS NULL) ORDER BY published_date DESC`
	rows, err := d.conn.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer closeRowsBOF(rows)

	var articles []Article
	for rows.Next() {
		var art Article
		err := rows.Scan(&art.GUID, &art.Title, &art.Link, &art.Description, &art.Content, &art.PublishedDate, &art.Score, &art.Analysis, &art.FeedURL, &art.Model, &art.Reported)
		if err != nil {
			return nil, err
		}
		articles = append(articles, art)
	}
	return articles, nil
}

// MarkArticlesReported marks the specified articles as reported.
func (d *DB) MarkArticlesReported(guids []string) error {
	if len(guids) == 0 {
		return nil
	}

	// Create placeholders and args for the SQL query.
	//   placeholders is the list of ?
	//   args is the list of guids
	placeholders := make([]string, len(guids))
	args := make([]interface{}, len(guids))
	for i, id := range guids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "UPDATE articles SET reported = 1 WHERE guid IN (" + strings.Join(placeholders, ",") + ")"
	_, err := d.conn.Exec(query, args...)
	return err
}

// ResetReported resets the reported flag for articles matching the title pattern.
func (d *DB) ResetReported(pattern string) (int64, error) {
	// Convert glob to SQL LIKE
	likePattern := strings.ReplaceAll(pattern, "*", "%")
	likePattern = strings.ReplaceAll(likePattern, "?", "_")

	query := "UPDATE articles SET reported = 0 WHERE title LIKE ?"
	res, err := d.conn.Exec(query, likePattern)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ResetReportedArticles resets the reported flag for the specified articles.
func (d *DB) ResetReportedArticles(guids []string) error {
	if len(guids) == 0 {
		return nil
	}
	placeholders := make([]string, len(guids))
	args := make([]interface{}, len(guids))
	for i, id := range guids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "UPDATE articles SET reported = 0 WHERE guid IN (" + strings.Join(placeholders, ",") + ")"
	_, err := d.conn.Exec(query, args...)
	return err
}

// ClearScores clears the score and analysis for the specified articles, effectively marking them for rescoring.
func (d *DB) ClearScores(guids []string) error {
	if len(guids) == 0 {
		return nil
	}
	placeholders := make([]string, len(guids))
	args := make([]interface{}, len(guids))
	for i, id := range guids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "UPDATE articles SET score = '', analysis = '', model = '' WHERE guid IN (" + strings.Join(placeholders, ",") + ")"
	_, err := d.conn.Exec(query, args...)
	return err
}

// Close closes the database connection.
func (d *DB) Close() {
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}
