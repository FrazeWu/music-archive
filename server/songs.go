package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SongInput represents song creation request
type SongInput struct {
	Title       string `form:"title" binding:"required"`
	Artist      string `form:"artist" binding:"required"`
	Album       string `form:"album"`
	ReleaseDate string `form:"release_date"` // YYYY-MM-DD
	TrackNumber int    `form:"track_number"`
	Lyrics      string `form:"lyrics"`
}

// SetupSongRoutes configures song-related routes
func SetupSongRoutes(router *gin.Engine, db *gorm.DB, s3Client *s3.S3) {
	songs := router.Group("/api/songs")
	{
		songs.GET("", GetSongsHandler(db))
		songs.GET("/:id", GetSongHandler(db))
		songs.POST("", AuthMiddleware(), CreateSongHandler(db, s3Client))
	}
}

// GetSongsHandler retrieves all approved songs
func GetSongsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []Song

		if err := db.Where("status = ?", "approved").
			Preload("Album").
			Preload("Album.Artist").
			Preload("Artists").
			Order("release_date ASC, track_number ASC, title ASC").
			Find(&songs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
			return
		}

		response := make([]map[string]interface{}, len(songs))
		for i, song := range songs {
			artistName := "Unknown Artist"
			albumTitle := "Unknown Album"
			albumYear := 0
			releaseDate := song.ReleaseDate.Format("2006-01-02")
			coverURL := song.CoverURL

			if song.Album != nil {
				albumTitle = song.Album.Title
				albumYear = song.Album.Year
				if song.Album.CoverURL != "" {
					coverURL = song.Album.CoverURL
				}
				if song.Album.Artist.Name != "" {
					artistName = song.Album.Artist.Name
				}
			}

			response[i] = map[string]interface{}{
				"id":           song.ID,
				"title":        song.Title,
				"artist":       artistName,
				"album":        albumTitle,
				"year":         albumYear,
				"release_date": releaseDate,
				"lyrics":       song.Lyrics,
				"audio_url":    song.AudioURL,
				"cover_url":    coverURL,
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetSongHandler retrieves a single song by ID
func GetSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
			return
		}

		var song Song
		if err := db.Preload("Album").Preload("Album.Artist").Preload("Artists").First(&song, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		artistName := "Unknown Artist"
		albumTitle := "Unknown Album"
		albumYear := 0
		releaseDate := song.ReleaseDate.Format("2006-01-02")
		coverURL := song.CoverURL

		if song.Album != nil {
			albumTitle = song.Album.Title
			albumYear = song.Album.Year
			if song.Album.CoverURL != "" {
				coverURL = song.Album.CoverURL
			}
			if song.Album.Artist.Name != "" {
				artistName = song.Album.Artist.Name
			}
		}

		response := map[string]interface{}{
			"id":           song.ID,
			"title":        song.Title,
			"artist":       artistName,
			"album":        albumTitle,
			"year":         albumYear,
			"release_date": releaseDate,
			"lyrics":       song.Lyrics,
			"audio_url":    song.AudioURL,
			"cover_url":    coverURL,
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateSongHandler creates a new song with optional audio upload
func CreateSongHandler(db *gorm.DB, s3Client *s3.S3) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SongInput

		if err := c.ShouldBind(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse ReleaseDate
		var releaseDate time.Time
		var err error
		if input.ReleaseDate != "" {
			releaseDate, err = time.Parse("2006-01-02", input.ReleaseDate)
			if err != nil {
				// Fallback to default or handle error. For now, use current time if invalid
				releaseDate = time.Now()
			}
		} else {
			releaseDate = time.Now()
		}

		// Handle File Upload and S3 Logic
		var audioURL string
		var coverURL string

		file, header, err := c.Request.FormFile("audio")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
			return
		}
		defer file.Close()

		safeArtist := strings.ReplaceAll(input.Artist, "/", "-")
		if safeArtist == "" {
			safeArtist = "Unknown Artist"
		}
		safeAlbum := strings.ReplaceAll(input.Album, "/", "-")
		if safeAlbum == "" {
			safeAlbum = "Unknown Album"
		}

		key := "music/" + safeArtist + "/" + safeAlbum + "/" + header.Filename

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(key),
			Body:   file,
			ACL:    aws.String("public-read"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
			return
		}

		audioURL = os.Getenv("S3_URL_PREFIX") + "/" + key

		coverFile, coverHeader, err := c.Request.FormFile("cover")
		if err == nil {
			defer coverFile.Close()

			coverKey := "music/" + safeArtist + "/" + safeAlbum + "/cover_" + coverHeader.Filename

			_, err = s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(coverKey),
				Body:   coverFile,
				ACL:    aws.String("public-read"),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
				return
			}

			coverURL = os.Getenv("S3_URL_PREFIX") + "/" + coverKey
		}

		// Transaction to ensure atomicity
		tx := db.Begin()

		// 1. Find or Create Artist
		var artist Artist
		if err := tx.FirstOrCreate(&artist, Artist{Name: input.Artist}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process artist"})
			return
		}

		// 2. Find or Create Album
		var album Album
		albumTitle := input.Album
		if albumTitle == "" {
			albumTitle = "Unknown Album"
		}

		// Ensure album is linked to artist
		if err := tx.FirstOrCreate(&album, Album{Title: albumTitle, ArtistID: artist.ID, Year: releaseDate.Year()}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process album"})
			return
		}

		if coverURL != "" && album.CoverURL == "" {
			album.CoverURL = coverURL
			if err := tx.Save(&album).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album cover"})
				return
			}
		}

		// 3. Create Song
		// Get User ID from context
		var userID *uint
		if idVal, exists := c.Get("user_id"); exists {
			// Convert float64 (from JWT JSON) to uint
			if fID, ok := idVal.(float64); ok {
				uID := uint(fID)
				userID = &uID
			} else if uID, ok := idVal.(uint); ok {
				userID = &uID
			}
		}

		song := Song{
			Title:       input.Title,
			ReleaseDate: releaseDate,
			TrackNumber: input.TrackNumber,
			Lyrics:      input.Lyrics,
			AudioURL:    audioURL,
			Status:      "pending",
			AlbumID:     &album.ID,
			UploadedBy:  userID,
		}

		if err := tx.Create(&song).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create song"})
			return
		}

		// 4. Associate Song with Artist (Many-to-Many)
		if err := tx.Model(&song).Association("Artists").Append(&artist); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate artist"})
			return
		}

		tx.Commit()

		// Reload song with associations for response
		db.Preload("Album").Preload("Artists").First(&song, song.ID)
		c.JSON(http.StatusCreated, song)
	}
}
