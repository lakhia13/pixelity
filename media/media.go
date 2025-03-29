package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"nas-app/models"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Buffer pool for processing images
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

const (
	MaxFileSize     = 32 << 20 // 32MB
	UploadDir       = "./uploads"
	ThumbnailDir    = "./thumbnails"
	ThumbnailWidth  = 300
	ThumbnailHeight = 300
)

var (
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	AllowedVideoTypes = map[string]bool{
		"video/mp4":  true,
		"video/webm": true,
		"video/avi":  true,
		"video/mpeg": true,
	}
)

// MediaService handles media-related operations
type MediaService struct {
	db *gorm.DB
}

// NewMediaService creates a new media service instance
func NewMediaService(db *gorm.DB) *MediaService {
	// Create necessary directories if they don't exist
	for _, dir := range []string{UploadDir, ThumbnailDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
		}
	}
	return &MediaService{db: db}
}

// ProcessUpload handles file upload and creates media item record
func (s *MediaService) ProcessUpload(file *multipart.FileHeader, userID uint) (*models.MediaItem, error) {
	// Get buffer from pool
	buf := bufferPool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	// Validate file size
	if file.Size > MaxFileSize {
		return nil, errors.New("file size exceeds maximum limit")
	}

	// Validate file type
	mimeType := file.Header.Get("Content-Type")
	mediaType := s.determineMediaType(mimeType)
	if mediaType == "" {
		return nil, errors.New("unsupported file type")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), generateRandomString(8), ext)
	filePath := filepath.Join(UploadDir, uniqueName)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("error saving file: %v", err)
	}

	// Generate thumbnail if it's an image
	var thumbnailPath string
	if mediaType == "image" {
		thumbnailPath, err = s.createThumbnail(filePath, uniqueName)
		if err != nil {
			os.Remove(filePath) // Clean up on error
			return nil, fmt.Errorf("error creating thumbnail: %v", err)
		}
	}

	// Create metadata
	metadata := map[string]interface{}{
		"originalName": file.Filename,
		"uploadTime":   time.Now(),
		"fileType":     mediaType,
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Create media item record
	mediaItem := &models.MediaItem{
		Filename:      uniqueName,
		Path:          filePath,
		ThumbnailPath: thumbnailPath,
		Size:          file.Size,
		MimeType:      mimeType,
		Type:          mediaType,
		UserID:        userID,
		Metadata:      string(metadataJSON),
	}

	if err := s.db.Create(mediaItem).Error; err != nil {
		os.Remove(filePath)
		if thumbnailPath != "" {
			os.Remove(thumbnailPath)
		}
		return nil, fmt.Errorf("error saving media record: %v", err)
	}

	return mediaItem, nil
}

// GetMediaItem retrieves a media item by ID
func (s *MediaService) GetMediaItem(id, userID uint) (*models.MediaItem, error) {
	var mediaItem models.MediaItem
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&mediaItem).Error; err != nil {
		return nil, err
	}
	return &mediaItem, nil
}

// DeleteMediaItem deletes a media item and its associated files
func (s *MediaService) DeleteMediaItem(id, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get media item
		var mediaItem models.MediaItem
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&mediaItem).Error; err != nil {
			return err
		}

		// Remove files
		if err := os.Remove(mediaItem.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing file: %v", err)
		}
		if mediaItem.ThumbnailPath != "" {
			if err := os.Remove(mediaItem.ThumbnailPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("error removing thumbnail: %v", err)
			}
		}

		// Remove album associations
		if err := tx.Exec("DELETE FROM album_media_items WHERE media_item_id = ?", id).Error; err != nil {
			return fmt.Errorf("error removing album associations: %v", err)
		}

		// Delete record
		if err := tx.Delete(&mediaItem).Error; err != nil {
			return fmt.Errorf("error deleting media record: %v", err)
		}

		return nil
	})
}

// UpdateMediaMetadata updates media item metadata
func (s *MediaService) UpdateMediaMetadata(id, userID uint, metadata map[string]interface{}) error {
	// Get existing media item
	var mediaItem models.MediaItem
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&mediaItem).Error; err != nil {
		return err
	}

	// Parse existing metadata
	var existingMetadata map[string]interface{}
	if err := json.Unmarshal([]byte(mediaItem.Metadata), &existingMetadata); err != nil {
		existingMetadata = make(map[string]interface{})
	}

	// Merge new metadata
	for k, v := range metadata {
		existingMetadata[k] = v
	}

	// Convert back to JSON
	metadataJSON, err := json.Marshal(existingMetadata)
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}

	// Update record
	return s.db.Model(&mediaItem).Update("metadata", string(metadataJSON)).Error
}

// ListUserMedia lists all media items for a user
func (s *MediaService) ListUserMedia(userID uint, mediaType string, offset, limit int) ([]models.MediaItem, int64, error) {
	var total int64
	query := s.db.Model(&models.MediaItem{}).Where("user_id = ?", userID)

	if mediaType != "" {
		query = query.Where("type = ?", mediaType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var items []models.MediaItem
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// createThumbnail generates a thumbnail for an image file
func (s *MediaService) createThumbnail(srcPath, filename string) (string, error) {
	// Open source image
	src, err := imaging.Open(srcPath)
	if err != nil {
		return "", err
	}

	// Calculate thumbnail dimensions while maintaining aspect ratio
	bounds := src.Bounds()
	ratio := float64(bounds.Dx()) / float64(bounds.Dy())
	var width, height int
	if ratio > 1 {
		width = ThumbnailWidth
		height = int(float64(ThumbnailWidth) / ratio)
	} else {
		height = ThumbnailHeight
		width = int(float64(ThumbnailHeight) * ratio)
	}

	// Resize image
	thumbnail := imaging.Resize(src, width, height, imaging.Lanczos)

	// Save thumbnail
	thumbnailPath := filepath.Join(ThumbnailDir, "thumb_"+filename)
	err = imaging.Save(thumbnail, thumbnailPath)
	if err != nil {
		return "", err
	}

	return thumbnailPath, nil
}

// determineMediaType returns the general type of media based on MIME type
func (s *MediaService) determineMediaType(mimeType string) string {
	if AllowedImageTypes[mimeType] {
		return "image"
	}
	if AllowedVideoTypes[mimeType] {
		return "video"
	}
	return ""
}

// generateRandomString generates a random string for unique filenames
func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

// ValidateMimeType checks if the given MIME type is allowed
func ValidateMimeType(mimeType string) bool {
	return AllowedImageTypes[mimeType] || AllowedVideoTypes[mimeType]
}

// GetMediaMetadata extracts metadata from a media item
func (s *MediaService) GetMediaMetadata(id, userID uint) (map[string]interface{}, error) {
	var mediaItem models.MediaItem
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&mediaItem).Error; err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(mediaItem.Metadata), &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// AddMediaTags adds tags to a media item's metadata
func (s *MediaService) AddMediaTags(id, userID uint, tags []string) error {
	metadata, err := s.GetMediaMetadata(id, userID)
	if err != nil {
		return err
	}

	// Get existing tags or create new tag slice
	existingTags, ok := metadata["tags"].([]interface{})
	if !ok {
		existingTags = make([]interface{}, 0)
	}

	// Convert existing tags to map for deduplication
	tagMap := make(map[string]bool)
	for _, tag := range existingTags {
		if tagStr, ok := tag.(string); ok {
			tagMap[tagStr] = true
		}
	}

	// Add new tags
	for _, tag := range tags {
		if !tagMap[tag] {
			existingTags = append(existingTags, tag)
			tagMap[tag] = true
		}
	}

	// Update metadata with new tags
	metadata["tags"] = existingTags
	return s.UpdateMediaMetadata(id, userID, metadata)
}

// RemoveMediaTags removes tags from a media item's metadata
func (s *MediaService) RemoveMediaTags(id, userID uint, tags []string) error {
	metadata, err := s.GetMediaMetadata(id, userID)
	if err != nil {
		return err
	}

	// Get existing tags
	existingTags, ok := metadata["tags"].([]interface{})
	if !ok {
		return nil // No tags to remove
	}

	// Convert tags to remove to map for quick lookup
	removeMap := make(map[string]bool)
	for _, tag := range tags {
		removeMap[tag] = true
	}

	// Filter out tags to remove
	newTags := make([]interface{}, 0)
	for _, tag := range existingTags {
		if tagStr, ok := tag.(string); ok && !removeMap[tagStr] {
			newTags = append(newTags, tagStr)
		}
	}

	// Update metadata with filtered tags
	metadata["tags"] = newTags
	return s.UpdateMediaMetadata(id, userID, metadata)
}