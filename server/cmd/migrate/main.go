package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Username  string    `gorm:"unique;not null;column:username"`
	Email     string    `gorm:"unique;not null;column:email"`
	Password  string    `gorm:"not null;column:password"`
	Role      string    `gorm:"default:'user';column:role"`
	CreatedAt time.Time `gorm:"column:createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt"`
}

func (User) TableName() string {
	return "Users"
}

type Artist struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null"`
	Bio       string `gorm:"type:text"`
	ImageURL  string
	CreatedAt time.Time `gorm:"column:createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt"`
}

func (Artist) TableName() string {
	return "Artists"
}

type Album struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Year      int
	CoverURL  string
	ArtistID  uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"column:createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt"`
}

func (Album) TableName() string {
	return "Albums"
}

type Song struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"not null"`
	ReleaseDate time.Time `gorm:"type:date"`
	TrackNumber int
	Lyrics      string `gorm:"type:text"`
	AudioURL    string `gorm:"not null"`
	CoverURL    string
	Status      string `gorm:"default:'pending'"`
	AlbumID     *uint
	UploadedBy  *uint
	CreatedAt   time.Time `gorm:"column:createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt"`
}

func (Song) TableName() string {
	return "Songs"
}

type SongArtist struct {
	SongID   uint `gorm:"primaryKey"`
	ArtistID uint `gorm:"primaryKey"`
}

func (SongArtist) TableName() string {
	return "song_artists"
}

type Correction struct {
	ID             uint   `gorm:"primaryKey"`
	SongID         uint   `gorm:"not null"`
	FieldName      string `gorm:"not null"`
	CurrentValue   string `gorm:"type:text"`
	CorrectedValue string `gorm:"type:text;not null"`
	Reason         string `gorm:"type:text"`
	UserID         *uint
	Status         string `gorm:"default:pending"`
	CreatedAt      time.Time
}

func main() {
	log.Println("Starting SQLite to MySQL migration...")

	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "../../database.sqlite"
	}

	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
			log.Fatal("MySQL connection details not provided. Set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME or MYSQL_DSN")
		}

		if dbPort == "" {
			dbPort = "3306"
		}

		mysqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	log.Printf("Connecting to SQLite: %s", sqlitePath)
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}
	log.Println("SQLite connected")

	log.Printf("Connecting to MySQL...")
	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	log.Println("MySQL connected")

	log.Println("Creating MySQL schema...")
	if err := mysqlDB.AutoMigrate(&User{}, &Artist{}, &Album{}, &Song{}, &SongArtist{}, &Correction{}); err != nil {
		log.Fatal("Failed to create MySQL schema:", err)
	}
	log.Println("MySQL schema created")

	log.Println("Migrating Users...")
	var users []User
	if err := sqliteDB.Find(&users).Error; err != nil {
		log.Fatal("Failed to read users from SQLite:", err)
	}
	if len(users) > 0 {
		if err := mysqlDB.Create(&users).Error; err != nil {
			log.Printf("Warning: Failed to migrate some users: %v", err)
		} else {
			log.Printf("Migrated %d users", len(users))
		}
	}

	log.Println("Migrating Artists...")
	var artists []Artist
	if err := sqliteDB.Find(&artists).Error; err != nil {
		log.Fatal("Failed to read artists from SQLite:", err)
	}
	if len(artists) > 0 {
		if err := mysqlDB.Create(&artists).Error; err != nil {
			log.Printf("Warning: Failed to migrate some artists: %v", err)
		} else {
			log.Printf("Migrated %d artists", len(artists))
		}
	}

	log.Println("Migrating Albums...")
	var albums []Album
	if err := sqliteDB.Find(&albums).Error; err != nil {
		log.Fatal("Failed to read albums from SQLite:", err)
	}
	if len(albums) > 0 {
		if err := mysqlDB.Create(&albums).Error; err != nil {
			log.Printf("Warning: Failed to migrate some albums: %v", err)
		} else {
			log.Printf("Migrated %d albums", len(albums))
		}
	}

	log.Println("Migrating Songs...")
	var songs []Song
	if err := sqliteDB.Find(&songs).Error; err != nil {
		log.Fatal("Failed to read songs from SQLite:", err)
	}
	if len(songs) > 0 {
		if err := mysqlDB.Create(&songs).Error; err != nil {
			log.Printf("Warning: Failed to migrate some songs: %v", err)
		} else {
			log.Printf("Migrated %d songs", len(songs))
		}
	}

	log.Println("Migrating Song-Artist relationships...")
	var songArtists []SongArtist
	if err := sqliteDB.Table("song_artists").Find(&songArtists).Error; err != nil {
		log.Printf("Warning: Failed to read song_artists: %v", err)
	} else if len(songArtists) > 0 {
		if err := mysqlDB.Table("song_artists").Create(&songArtists).Error; err != nil {
			log.Printf("Warning: Failed to migrate song_artists: %v", err)
		} else {
			log.Printf("Migrated %d song-artist relationships", len(songArtists))
		}
	}

	log.Println("Migrating Corrections...")
	var corrections []Correction
	if err := sqliteDB.Find(&corrections).Error; err != nil {
		log.Fatal("Failed to read corrections from SQLite:", err)
	}
	if len(corrections) > 0 {
		if err := mysqlDB.Create(&corrections).Error; err != nil {
			log.Printf("Warning: Failed to migrate some corrections: %v", err)
		} else {
			log.Printf("Migrated %d corrections", len(corrections))
		}
	}

	log.Println("Migration completed successfully!")
}
