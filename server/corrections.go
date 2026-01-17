package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CorrectionInput represents correction submission request
type CorrectionInput struct {
	SongID         uint   `json:"song_id" binding:"required"`
	FieldName      string `json:"field_name" binding:"required"`
	CurrentValue   string `json:"current_value"`
	CorrectedValue string `json:"corrected_value" binding:"required"`
	Reason         string `json:"reason"`
}

// SetupCorrectionRoutes configures correction-related routes
func SetupCorrectionRoutes(router *gin.Engine, db *gorm.DB) {
	corrections := router.Group("/api/corrections")
	{
		corrections.POST("", AuthMiddleware(), CreateCorrectionHandler(db))
		// TODO: Add routes for listing, approving, rejecting corrections
	}
}

// CreateCorrectionHandler creates a new correction submission
func CreateCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var input CorrectionInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		correction := Correction{
			SongID:         input.SongID,
			FieldName:      input.FieldName,
			CurrentValue:   input.CurrentValue,
			CorrectedValue: input.CorrectedValue,
			Reason:         input.Reason,
			UserID:         &[]uint{userID.(uint)}[0],
		}

		if err := db.Create(&correction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Correction submitted successfully", "id": correction.ID})
	}
}
