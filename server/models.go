package main

import (
	"time"
)

// User represents a user account
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;column:id"`
	Username  string    `json:"username" gorm:"unique;not null;column:username"`
	Email     string    `json:"email" gorm:"unique;not null;column:email"`
	Password  string    `json:"-" gorm:"not null;column:password"`
	Role      string    `json:"role" gorm:"default:'user';column:role"` // 'user' or 'admin'
	CreatedAt time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updatedAt"`
	Songs     []Song    `json:"songs,omitempty" gorm:"foreignKey:UploadedBy"`
}

// TableName overrides the table name used by User to `Users`
func (User) TableName() string {
	return "Users"
}

// Artist represents a music artist
type Artist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"unique;not null"`
	Bio       string    `json:"bio" gorm:"type:text"`
	ImageURL  string    `json:"image_url"`
	Albums    []Album   `json:"albums,omitempty" gorm:"foreignKey:ArtistID"`
	Songs     []Song    `json:"songs,omitempty" gorm:"many2many:song_artists;"`
	CreatedAt time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updatedAt"`
}

// TableName overrides the table name used by Artist to `Artists`
func (Artist) TableName() string {
	return "Artists"
}

// Album represents a music album
type Album struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Year      int       `json:"year"`
	CoverURL  string    `json:"cover_url"`
	ArtistID  uint      `json:"artist_id" gorm:"not null"`
	Artist    Artist    `json:"artist,omitempty"`
	Songs     []Song    `json:"songs,omitempty" gorm:"foreignKey:AlbumID"`
	CreatedAt time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updatedAt"`
}

// TableName overrides the table name used by Album to `Albums`
func (Album) TableName() string {
	return "Albums"
}

// Song represents a music track
type Song struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	ReleaseDate time.Time `json:"release_date" gorm:"type:date"`
	TrackNumber int       `json:"track_number"`
	Lyrics      string    `json:"lyrics" gorm:"type:text"`
	AudioURL    string    `json:"audio_url" gorm:"not null"`
	CoverURL    string    `json:"cover_url"`
	Status      string    `json:"status" gorm:"default:'pending'"` // 'pending', 'approved', 'rejected'
	AlbumID     *uint     `json:"album_id"`
	Album       *Album    `json:"album,omitempty"`
	Artists     []Artist  `json:"artists,omitempty" gorm:"many2many:song_artists;"`
	UploadedBy  *uint     `json:"uploaded_by"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UploadedBy"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:createdAt"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updatedAt"`
}

// TableName overrides the table name used by Song to `Songs`
func (Song) TableName() string {
	return "Songs"
}

// Correction represents a user-submitted correction for a song
type Correction struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SongID         uint      `json:"song_id" gorm:"not null"`
	FieldName      string    `json:"field_name" gorm:"not null"`
	CurrentValue   string    `json:"current_value" gorm:"type:text"`
	CorrectedValue string    `json:"corrected_value" gorm:"type:text;not null"`
	Reason         string    `json:"reason" gorm:"type:text"`
	UserID         *uint     `json:"user_id"`
	Status         string    `json:"status" gorm:"default:pending"`
	CreatedAt      time.Time `json:"created_at"`
	Song           Song      `json:"-" gorm:"foreignKey:SongID"`
	User           *User     `json:"-" gorm:"foreignKey:UserID"`
}
