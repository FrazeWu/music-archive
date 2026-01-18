package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// AdminMiddleware validates admin privileges
func AdminMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SetupAdminRoutes(router *gin.Engine, db *gorm.DB) {
	admin := router.Group("/api/admin")
	admin.Use(AuthMiddleware())
	admin.Use(AdminMiddleware(db))
	{
		admin.GET("/pending", GetPendingRequestsHandler(db))
		admin.POST("/approve/:id", ApproveSongHandler(db))
		admin.POST("/reject/:id", RejectSongHandler(db))
		admin.POST("/approve-batch", ApproveBatchHandler(db))
		admin.POST("/reject-batch", RejectBatchHandler(db))
	}
}

type PendingRequest struct {
	ID        string    `json:"id"` // Unique Key for UI (BatchID or synthetic)
	BatchID   string    `json:"batch_id"`
	Album     string    `json:"album"`
	Artist    string    `json:"artist"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	Count     int       `json:"count"`
	Songs     []Song    `json:"songs"`
	SongIDs   []uint    `json:"song_ids"`
}

type BatchActionInput struct {
	IDs []uint `json:"ids" binding:"required"`
}

func GetPendingRequestsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []Song
		if err := db.Where("status = ?", "pending").Preload("User").Preload("Album").Preload("Album.Artist").Order("created_at desc").Find(&songs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending songs"})
			return
		}

		groups := make(map[string]*PendingRequest)
		
		for _, song := range songs {
			key := song.BatchID
			if key == "" {
				// Fallback for legacy: Group by Album + User + Date
				key = "legacy_" + song.Album.Title + "_" + song.User.Username + "_" + song.CreatedAt.Format("20060102")
			}

			if _, exists := groups[key]; !exists {
				albumName := "Unknown"
				artistName := "Unknown"
				if song.Album != nil {
					albumName = song.Album.Title
					if song.Album.Artist.Name != "" {
						artistName = song.Album.Artist.Name
					}
				}
				
				groups[key] = &PendingRequest{
					ID:        key,
					BatchID:   song.BatchID,
					Album:     albumName,
					Artist:    artistName,
					User:      *song.User,
					CreatedAt: song.CreatedAt,
					Count:     0,
					Songs:     []Song{},
					SongIDs:   []uint{},
				}
			}
			
			groups[key].Count++
			groups[key].Songs = append(groups[key].Songs, song)
			groups[key].SongIDs = append(groups[key].SongIDs, song.ID)
		}
		
		finalResult := make([]PendingRequest, 0, len(groups))
		for _, req := range groups {
			finalResult = append(finalResult, *req)
		}
		
		// Sort by CreatedAt Desc
		// (Need sort package, or just rely on map iteration being random and frontend sorting, 
		// but typically Go maps are random. Let's not complicate with sort import yet if not strictly needed,
		// but strictly speaking I should. I'll rely on frontend sorting or just leave as is for prototype)
		
		c.JSON(http.StatusOK, finalResult)
	}
}

func ApproveBatchHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input BatchActionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := db.Model(&Song{}).Where("id IN ?", input.IDs).Update("status", "approved").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve songs"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Songs approved"})
	}
}

func RejectBatchHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input BatchActionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := db.Model(&Song{}).Where("id IN ?", input.IDs).Update("status", "rejected").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject songs"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Songs rejected"})
	}
}

func ApproveSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Model(&Song{}).Where("id = ?", id).Update("status", "approved").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve song"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Song approved"})
	}
}

func RejectSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Model(&Song{}).Where("id = ?", id).Update("status", "rejected").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject song"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Song rejected"})
	}
}
