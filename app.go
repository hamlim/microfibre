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

// Media struct to represent images or videos
type Media struct {
	Type  string `json:"type"`  // "image" or "video"
	URL   string `json:"url"`   // Source URL for the media
	Title string `json:"title"` // Optional title for the media
}

// Updated Post struct with media field
type Post struct {
	ID              int        `json:"id"`
	Body            string     `json:"body"`
	UpdatedTime     *time.Time `json:"updated_time"`
	UpdatedTimezone string     `json:"updated_timezone"`
	CreatedTime     time.Time  `json:"created_time"`
	CreatedTimezone string     `json:"created_timezone"`
	Location        *string    `json:"location"`
	Media           []Media    `json:"media"`
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
	CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    body TEXT,
    updated_time DATETIME,
    updated_timezone TEXT NULL,
    created_time DATETIME,
    created_timezone TEXT,
    location TEXT
);

CREATE TABLE IF NOT EXISTS media (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id INTEGER,
	type TEXT,
	url TEXT,
	title TEXT,
	FOREIGN KEY (post_id) REFERENCES posts (id)
);
    `)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database tables created successfully")

	r.GET("/v1/posts", func(c *gin.Context) {
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
		query := "SELECT id, body, updated_time, updated_timezone, created_time, created_timezone, location FROM posts ORDER BY created_time DESC LIMIT ? OFFSET ?"
		rows, err := db.Query(query, pageSizeInt, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			var updatedTime, createdTime sql.NullTime
			var updatedTimezone, createdTimezone sql.NullString
			var location sql.NullString

			if err := rows.Scan(&post.ID, &post.Body, &updatedTime, &updatedTimezone, &createdTime, &createdTimezone, &location); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if updatedTime.Valid {
				post.UpdatedTime = &updatedTime.Time
				post.UpdatedTimezone = updatedTimezone.String
			}
			post.CreatedTime = createdTime.Time
			post.CreatedTimezone = createdTimezone.String
			if location.Valid {
				post.Location = &location.String
			}

			// Fetch media for the current post
			mediaRows, err := db.Query("SELECT type, url, title FROM media WHERE post_id = ?", post.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer mediaRows.Close()

			var media []Media
			for mediaRows.Next() {
				var m Media
				if err := mediaRows.Scan(&m.Type, &m.URL, &m.Title); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				media = append(media, m)
			}

			post.Media = media
			posts = append(posts, post)
		}

		totalCount, err := getTotalRecordCount(db)
		if err != nil {
			log.Printf("Failed to calculate the total number of records!")
		}

		c.Writer.Header().Set("X-Total-Count", strconv.Itoa(totalCount))

		c.JSON(http.StatusOK, posts)
	})

	r.POST("/v1/post", func(c *gin.Context) {
		// Parse the JSON payload
		var post Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure that the 'body' field is provided
		if post.Body == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The 'body' field is required."})
			return
		}

		// Set the 'created_time' to the current time with the provided timezone
		post.CreatedTime = time.Now()

		if post.Location == nil {
			post.Location = nil // You can omit this line, as it's already nil by default
		}

		// Insert the new post into the database
		result, err := db.Exec(`
        INSERT INTO posts (body, updated_time, updated_timezone, created_time, created_timezone, location)
        VALUES (?, NULL, NULL, ?, ?, ?)
    `, post.Body, post.CreatedTime, post.CreatedTimezone, post.Location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the ID of the newly inserted post
		postID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert media entries into the database
		for _, media := range post.Media {
			_, err := db.Exec(`
            INSERT INTO media (post_id, type, url, title)
            VALUES (?, ?, ?, ?)
        `, postID, media.Type, media.URL, media.Title)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Post created",
			"id":      postID,
		})
	})

	r.POST("/v1/update", func(c *gin.Context) {
		// Parse the payload from the request body
		var updatePayload gin.H
		if err := c.ShouldBindJSON(&updatePayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure that the post ID is provided in the payload
		postID, ok := updatePayload["id"].(float64)
		if !ok || postID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The 'id' field is required in the payload."})
			return
		}

		// Retrieve the existing post entry from the database using the post ID
		var existingPost Post
		err := db.QueryRow("SELECT id, body, updated_time, created_time, created_timezone, location FROM posts WHERE id = ?", int(postID)).
			Scan(&existingPost.ID, &existingPost.Body, &existingPost.UpdatedTime, &existingPost.CreatedTime, &existingPost.CreatedTimezone, &existingPost.Location)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found or error in database query", "details": err.Error()})
			return
		}

		// Handle NULL value for updated_timezone
		var updatedTimezone sql.NullString
		err = db.QueryRow("SELECT updated_timezone FROM posts WHERE id = ?", int(postID)).Scan(&updatedTimezone)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving updated_timezone", "details": err.Error()})
			return
		}
		existingPost.UpdatedTimezone = updatedTimezone.String

		// Update the existing post entry with values from the payload
		// You may want to add checks to only update fields that are provided in the payload

		if body, ok := updatePayload["body"].(string); ok {
			existingPost.Body = body
		}

		if location, ok := updatePayload["location"].(string); ok {
			existingPost.Location = &location
		}

		if media, ok := updatePayload["media"]; ok {
			// Ensure that 'media' is not nil before attempting to convert
			if mediaArray, isArray := media.([]interface{}); isArray {
				existingPost.Media = make([]Media, len(mediaArray))
				for i, m := range mediaArray {
					mediaMap, ok := m.(map[string]interface{})
					if !ok {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media format in payload"})
						return
					}
					mediaType, _ := mediaMap["type"].(string)
					mediaURL, _ := mediaMap["url"].(string)
					mediaTitle, _ := mediaMap["title"].(string)
					existingPost.Media[i] = Media{
						Type:  mediaType,
						URL:   mediaURL,
						Title: mediaTitle,
					}
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media format in payload"})
				return
			}
		}

		// Set the 'updatedTime' and 'updatedTimezone' to the current time and provided timezone
		currentTime := time.Now()
		existingPost.UpdatedTime = &currentTime
		existingPost.UpdatedTimezone = updatePayload["updated_timezone"].(string)

		// Delete existing media entries associated with the post
		_, deleteErr := db.Exec(`
DELETE FROM media
WHERE post_id = ?
`, existingPost.ID)

		if deleteErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": deleteErr.Error()})
			return
		}

		// Insert new media entries from the update payload
		for _, media := range existingPost.Media {
			_, insertErr := db.Exec(`
INSERT INTO media (type, url, title, post_id)
VALUES (?, ?, ?, ?)
`, media.Type, media.URL, media.Title, existingPost.ID)

			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
				return
			}
		}

		// Now, update the post entry in the posts table
		_, updateErr := db.Exec(`
UPDATE posts
SET body = ?, updated_time = ?, updated_timezone = ?, location = ?
WHERE id = ?
`, existingPost.Body, existingPost.UpdatedTime, existingPost.UpdatedTimezone, existingPost.Location, existingPost.ID)

		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": updateErr.Error()})
			return
		}

		c.JSON(http.StatusOK, existingPost)
	})

	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, "<html><body><marquee>Hello World!</marquee></body></html>")
	})

	r.Run(host)
}

func getTotalRecordCount(db *sql.DB) (int, error) {
	query := "SELECT COUNT(*) FROM posts"
	var totalCount int
	err := db.QueryRow(query).Scan(&totalCount)
	if err != nil {
		return 0, err
	}
	return totalCount, nil
}
