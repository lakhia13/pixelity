package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"time"
)

func uploadMedia(c *gin.Context) {
	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Generate unique filename
	filename := filepath.Join("media", time.Now().Format("20060102150405") + "_" + file.Filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create media record in database
	media := Media{
		Filename: file.Filename,
		Path:     filename,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
		UserID:   c.GetUint("user_id"), // from JWT token
	}

	if err := db.Create(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": media.ID, "filename": media.Filename})
}

func getMedia(c *gin.Context) {
	id := c.Param("id")
	var media Media
	
	if err := db.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Check if user has access to this media
	if media.UserID != c.GetUint("user_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.File(media.Path)
}

func listMedia(c *gin.Context) {
	var media []Media
	userID := c.GetUint("user_id")

	if err := db.Where("user_id = ?", userID).Find(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch media"})
		return
	}

	c.JSON(http.StatusOK, media)
}

func createAlbum(c *gin.Context) {
	var album Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	album.UserID = c.GetUint("user_id")
	if err := db.Create(&album).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
		return
	}

	c.JSON(http.StatusOK, album)
}

func listAlbums(c *gin.Context) {
	var albums []Album
	userID := c.GetUint("user_id")

	if err := db.Where("user_id = ?", userID).Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
		return
	}

	c.JSON(http.StatusOK, albums)
}

func getAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	if err := db.Preload("Media").First(&album, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	if album.UserID != c.GetUint("user_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, album)
}