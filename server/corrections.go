package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CorrectionInput struct {
	SongID         uint   `json:"song_id" binding:"required"`
	FieldName      string `json:"field_name" binding:"required"`
	CurrentValue   string `json:"current_value"`
	CorrectedValue string `json:"corrected_value" binding:"required"`
	Reason         string `json:"reason"`
}

type BatchCorrectionInput struct {
	SongID      uint             `json:"song_id" binding:"required"`
	Artist      string           `json:"artist"`
	Album       string           `json:"album"`
	ReleaseDate string           `json:"release_date"`
	Reason      string           `json:"reason"`
	Corrections []CorrectionItem `json:"corrections"`
}

type CorrectionItem struct {
	FieldName      string `json:"field_name" binding:"required"`
	CurrentValue   string `json:"current_value"`
	CorrectedValue string `json:"corrected_value" binding:"required"`
}

func SetupCorrectionRoutes(router *gin.Engine, db *gorm.DB) {
	corrections := router.Group("/api/corrections")
	{
		corrections.POST("", AuthMiddleware(), CreateCorrectionHandler(db))
		corrections.POST("/batch", AuthMiddleware(), CreateBatchCorrectionHandler(db))
	}
}

func CreateCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var input CorrectionInput

		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("Correction input error: %v", err)
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
			log.Printf("Failed to create correction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit correction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Correction submitted successfully", "id": correction.ID})
	}
}

func CreateBatchCorrectionHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var input BatchCorrectionInput

		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("Batch correction input error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Received batch correction: song_id=%d, artist=%s, album=%s, release_date=%s, reason=%s",
			input.SongID, input.Artist, input.Album, input.ReleaseDate, input.Reason)

		var corrections []Correction

		if len(input.Corrections) > 0 {
			for _, item := range input.Corrections {
				if item.CorrectedValue != "" && item.CorrectedValue != item.CurrentValue {
					corrections = append(corrections, Correction{
						SongID:         input.SongID,
						FieldName:      item.FieldName,
						CurrentValue:   item.CurrentValue,
						CorrectedValue: item.CorrectedValue,
						Reason:         input.Reason,
						UserID:         &[]uint{userID.(uint)}[0],
					})
				}
			}
		} else {
			if input.Artist != "" {
				corrections = append(corrections, Correction{
					SongID:         input.SongID,
					FieldName:      "artist",
					CorrectedValue: input.Artist,
					Reason:         input.Reason,
					UserID:         &[]uint{userID.(uint)}[0],
				})
			}
			if input.Album != "" {
				corrections = append(corrections, Correction{
					SongID:         input.SongID,
					FieldName:      "album",
					CorrectedValue: input.Album,
					Reason:         input.Reason,
					UserID:         &[]uint{userID.(uint)}[0],
				})
			}
			if input.ReleaseDate != "" {
				corrections = append(corrections, Correction{
					SongID:         input.SongID,
					FieldName:      "release_date",
					CorrectedValue: input.ReleaseDate,
					Reason:         input.Reason,
					UserID:         &[]uint{userID.(uint)}[0],
				})
			}
		}

		if len(corrections) == 0 {
			log.Printf("No valid corrections provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供有效的修正内容"})
			return
		}

		log.Printf("Creating %d corrections", len(corrections))

		if err := db.Create(&corrections).Error; err != nil {
			log.Printf("Failed to create corrections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit corrections"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Corrections submitted successfully",
			"count":        len(corrections),
			"correction_ids": func() []uint {
				ids := make([]uint, len(corrections))
				for i, c := range corrections {
					ids[i] = c.ID
				}
				return ids
			}(),
		})
	}
}
