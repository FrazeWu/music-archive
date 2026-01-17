package main

import (
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
		admin.GET("/pending", GetPendingSongsHandler(db))
		admin.POST("/approve/:id", ApproveSongHandler(db))
		admin.POST("/reject/:id", RejectSongHandler(db))
	}
}

func GetPendingSongsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []Song
		if err := db.Where("status = ?", "pending").Preload("User").Find(&songs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending songs"})
			return
		}
		c.JSON(http.StatusOK, songs)
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
