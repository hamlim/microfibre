package main

import (
	"database/sql"
	"log"
	"os"

	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type StatusUpdate struct {
	ID          int        `json:"id"`
	Body        string     `json:"body" binding:"required"`
	UpdatedTime *time.Time `json:"updated_time"`
	CreatedTime time.Time  `json:"created_time" binding:"required"`
	Location    *string    `json:"location"`
}

func main() {

	r := gin.Default()

	// Get the database file path from the environment variable
	dbFilePath := os.Getenv("DB_FILE_PATH")
	if dbFilePath == "" {
		// Default path for local development
		dbFilePath = "./data/microfibre.db"
	}

	// Get the host from the environment
	host := os.Getenv("HOST")
	if host == "" {
		// Fallback to localhost:8080 for local dev
		host = "127.0.0.1:8080"
	}

	log.Printf("Started with env, host = %s, db = %s", host, dbFilePath)

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatal(err, dbFilePath)
	}
	defer db.Close()

	// Create a table for status updates
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS status_updates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		body TEXT,
		updated_time DATETIME,
		created_time DATETIME,
		location TEXT
)
    `)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database created successfully")

	r.GET("/status_updates", func(c *gin.Context) {
		// Implement the logic to retrieve status updates from the database
		rows, err := db.Query("SELECT id, body, updated_time, created_time, location FROM status_updates")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var statusUpdates []StatusUpdate
		for rows.Next() {
			var status StatusUpdate
			var updatedTime, createdTime sql.NullTime
			var location sql.NullString

			if err := rows.Scan(&status.ID, &status.Body, &updatedTime, &createdTime, &location); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if updatedTime.Valid {
				status.UpdatedTime = &updatedTime.Time
			}
			status.CreatedTime = createdTime.Time
			if location.Valid {
				status.Location = &location.String
			}

			statusUpdates = append(statusUpdates, status)
		}

		c.JSON(http.StatusOK, statusUpdates)
	})

	r.POST("/status_updates", func(c *gin.Context) {
		var status StatusUpdate
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if 'updated_time' and 'location' are nil and set to default values if they are
		if status.UpdatedTime == nil {
			defaultTime := time.Now()
			status.UpdatedTime = &defaultTime
		}
		if status.Location == nil {
			status.Location = nil // You can omit this line, as it's already nil by default
		}

		// Insert the new status update into the database
		result, err := db.Exec(`
            INSERT INTO status_updates (body, updated_time, created_time, location)
            VALUES (?, ?, ?, ?)
        `, status.Body, status.UpdatedTime, status.CreatedTime, status.Location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the ID of the newly inserted status update
		id, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Status update created",
			"id":      id,
		})
	})

	r.GET("/", func (c *gin.Context) {
		c.Header("Content-Type", "text/html");
		c.String(http.StatusOK, "<html><body><marquee>Hello World!</marquee></body></html>");
	})

	r.Run(host)
}
