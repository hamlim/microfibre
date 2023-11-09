package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"net/http"

	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type StatusUpdate struct {
	ID          int        `json:"id"`
	Body        string     `json:"body"`
	UpdatedTime *time.Time `json:"updated_time"`
	CreatedTime time.Time  `json:"created_time"`
	Location    *string    `json:"location"`
}

func TokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the 'secret-token' header is present in the request
		secretToken := c.GetHeader("secret-token")

		if secretToken == "" {
			// Return an error response for unauthorized requests
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized. Please open an issue if you'd like to be added as a supported client!"})
			c.Abort()
			return
		}

		apiVersion := c.GetHeader("api-version")

		if apiVersion != "v1" {
			// Return an error response for unauthorized requests
			c.JSON(http.StatusHTTPVersionNotSupported, gin.H{"error": "Unsupported API version requested."})
			c.Abort()
			return
		}

		// Continue processing the request if the header is present
		c.Next()
	}
}

func main() {

	r := gin.Default()

	// Create a CORS middleware instance with your desired configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	config.AllowCredentials = true

	// Use the configured middleware
	r.Use(cors.New(config))

	r.Use(TokenMiddleware())

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

	r.GET("/v1/read", func(c *gin.Context) {
		// Extract query parameters for pagination
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")

		// Parse the page and pageSize as integers
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}

		// Calculate the offset based on page and pageSize
		offset := (pageInt - 1) * pageSizeInt

		// Adjust your SQL query to retrieve a specific page of results
		query := "SELECT id, body, updated_time, created_time, location FROM status_updates LIMIT ? OFFSET ?"
		rows, err := db.Query(query, pageSizeInt, offset)
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

		totalCount, err := getTotalRecordCount(db)
		if err != nil {
			log.Printf("Failed to calculate the total number of records!")
		}

		c.Writer.Header().Set("X-Total-Count", strconv.Itoa(totalCount))

		c.JSON(http.StatusOK, statusUpdates)
	})

	r.POST("/v1/create", func(c *gin.Context) {
		status := StatusUpdate{}

		// Parse the JSON payload
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure that the 'body' field is provided
		if status.Body == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The 'body' field is required."})
			return
		}

		// Set the 'created_time' to the current time
		status.CreatedTime = time.Now()

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

	r.POST("/v1/update", func(c *gin.Context) {
		// Parse the status ID from the query parameter
		statusID := c.Query("id")

		// Retrieve the existing status entry from the database using the status ID
		var existingStatus StatusUpdate
		err := db.QueryRow("SELECT id, body, updated_time, created_time, location FROM status_updates WHERE id = ?", statusID).
			Scan(&existingStatus.ID, &existingStatus.Body, &existingStatus.UpdatedTime, &existingStatus.CreatedTime, &existingStatus.Location)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status not found or error in database query"})
			return
		}

		// Parse the payload from the request body
		var updatePayload StatusUpdate
		if err := c.ShouldBindJSON(&updatePayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update the existing status entry with values from the payload
		// You may want to add checks to only update fields that are provided in the payload

		if updatePayload.Body != "" {
			existingStatus.Body = updatePayload.Body
		}

		if updatePayload.Location != nil {
			existingStatus.Location = updatePayload.Location
		}

		// Set the 'updatedTime' to the current time
		updatedTime := time.Now()
		existingStatus.UpdatedTime = &updatedTime

		// Update the status entry in the database
		_, updateErr := db.Exec(`
				UPDATE status_updates
				SET body = ?, updated_time = ?, location = ?
				WHERE id = ?
		`, existingStatus.Body, existingStatus.UpdatedTime, existingStatus.Location, statusID)

		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": updateErr.Error()})
			return
		}

		c.JSON(http.StatusOK, existingStatus)
	})

	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, "<html><body><marquee>Hello World!</marquee></body></html>")
	})

	r.Run(host)
}

func getTotalRecordCount(db *sql.DB) (int, error) {
	query := "SELECT COUNT(*) FROM status_updates"
	var totalCount int
	err := db.QueryRow(query).Scan(&totalCount)
	if err != nil {
		return 0, err
	}
	return totalCount, nil
}
