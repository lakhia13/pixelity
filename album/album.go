package album

import (
	"errors"
	"gorm.io/gorm"
	"nas-app/models"
	"time"
)

// AlbumService handles album-related operations
type AlbumService struct {
	db *gorm.DB
}

// NewAlbumService creates a new album service instance
func NewAlbumService(db *gorm.DB) *AlbumService {
	return &AlbumService{db: db}
}

// CreateAlbum creates a new album
func (s *AlbumService) CreateAlbum(name, description string, userID uint) (*models.Album, error) {
	album := &models.Album{
		Name:        name,
		Description: description,
		UserID:      userID,
	}

	result := s.db.Create(album)
	if result.Error != nil {
		return nil, result.Error
	}

	return album, nil
}

// GetAlbum retrieves an album by ID and verifies ownership
func (s *AlbumService) GetAlbum(id, userID uint) (*models.Album, error) {
	var album models.Album
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&album).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("album not found or access denied")
		}
		return nil, err
	}

	// Get media count
	var count int64
	s.db.Model(&models.AlbumMediaItem{}).Where("album_id = ?", id).Count(&count)
	album.MediaCount = int(count)

	return &album, nil
}

// UpdateAlbum updates an album's details
func (s *AlbumService) UpdateAlbum(id uint, name, description string, userID uint) error {
	result := s.db.Model(&models.Album{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"name":        name,
			"description": description,
			"updated_at":  time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("album not found or access denied")
	}

	return nil
}

// DeleteAlbum deletes an album and its associations
func (s *AlbumService) DeleteAlbum(id, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify ownership
		var album models.Album
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&album).Error; err != nil {
			return errors.New("album not found or access denied")
		}

		// Delete album-media associations
		if err := tx.Delete(&models.AlbumMediaItem{}, "album_id = ?", id).Error; err != nil {
			return err
		}

		// Delete the album
		if err := tx.Delete(&album).Error; err != nil {
			return err
		}

		return nil
	})
}

// ListUserAlbums retrieves all albums for a user
func (s *AlbumService) ListUserAlbums(userID uint) ([]models.Album, error) {
	var albums []models.Album
	err := s.db.Where("user_id = ?", userID).Find(&albums).Error
	if err != nil {
		return nil, err
	}

	// Get media counts for each album
	for i := range albums {
		var count int64
		s.db.Model(&models.AlbumMediaItem{}).Where("album_id = ?", albums[i].ID).Count(&count)
		albums[i].MediaCount = int(count)
	}

	return albums, nil
}

// AddMediaToAlbum adds media items to an album
func (s *AlbumService) AddMediaToAlbum(albumID uint, mediaIDs []uint, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify album ownership
		var album models.Album
		if err := tx.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
			return errors.New("album not found or access denied")
		}

		// Get current max display order
		var maxOrder int
		tx.Model(&models.AlbumMediaItem{}).
			Where("album_id = ?", albumID).
			Select("COALESCE(MAX(display_order), 0)").
			Scan(&maxOrder)

		// Add media items
		for i, mediaID := range mediaIDs {
			// Verify media ownership using raw query
			var count int64
			if err := tx.Table("media_items").
				Where("id = ? AND user_id = ?", mediaID, userID).
				Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				return errors.New("media item not found or access denied")
			}

			// Create association
			item := models.AlbumMediaItem{
				AlbumID:      albumID,
				MediaItemID:  mediaID,
				DisplayOrder: maxOrder + i + 1,
			}

			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveMediaFromAlbum removes media items from an album
func (s *AlbumService) RemoveMediaFromAlbum(albumID uint, mediaID uint, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify album ownership
		var album models.Album
		if err := tx.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
			return errors.New("album not found or access denied")
		}

		// Remove association
		result := tx.Delete(&models.AlbumMediaItem{}, "album_id = ? AND media_item_id = ?", albumID, mediaID)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("media item not found in album")
		}

		return nil
	})
}

// MediaItemData represents the media item data structure for queries
type MediaItemData struct {
	ID            uint      `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt
	Filename      string
	Path          string
	ThumbnailPath string
	Size          int64
	MimeType      string
	Type          string
	UserID        uint
	Metadata      string
}

// GetAlbumMedia retrieves all media items in an album
func (s *AlbumService) GetAlbumMedia(albumID, userID uint) ([]MediaItemData, error) {
	// First verify album ownership
	var album models.Album
	if err := s.db.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
		return nil, errors.New("album not found or access denied")
	}

	var items []MediaItemData
	err := s.db.Raw(`
		SELECT m.* 
		FROM media_items m 
		INNER JOIN album_media_items ami ON m.id = ami.media_item_id 
		WHERE ami.album_id = ? 
		ORDER BY ami.display_order
	`, albumID).Scan(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// SetAlbumCover sets an album's cover image
func (s *AlbumService) SetAlbumCover(albumID, mediaID, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify album ownership
		var album models.Album
		if err := tx.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
			return errors.New("album not found or access denied")
		}

		// Get media path using raw query
		var thumbnailPath string
		err := tx.Table("media_items").
			Select("thumbnail_path").
			Where("id = ? AND user_id = ?", mediaID, userID).
			Row().
			Scan(&thumbnailPath)

		if err != nil {
			return errors.New("media item not found or access denied")
		}

		// Verify media item is in the album
		var count int64
		if err := tx.Model(&models.AlbumMediaItem{}).
			Where("album_id = ? AND media_item_id = ?", albumID, mediaID).
			Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			return errors.New("media item is not in the album")
		}

		// Update album cover
		return tx.Model(&album).Update("cover", thumbnailPath).Error
	})
}

// ReorderAlbumMedia updates the display order of media items in an album
func (s *AlbumService) ReorderAlbumMedia(albumID uint, mediaOrder []uint, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify album ownership
		var album models.Album
		if err := tx.Where("id = ? AND user_id = ?", albumID, userID).First(&album).Error; err != nil {
			return errors.New("album not found or access denied")
		}

		// Update order for each media item
		for order, mediaID := range mediaOrder {
			result := tx.Model(&models.AlbumMediaItem{}).
				Where("album_id = ? AND media_item_id = ?", albumID, mediaID).
				Update("display_order", order+1)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return errors.New("one or more media items not found in album")
			}
		}

		return nil
	})
}