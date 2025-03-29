package search

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"nas-app/models"
	"sort"
	"time"
)

// SearchService handles search operations
type SearchService struct {
	db *gorm.DB
}

// SearchFilter defines parameters for searching
type SearchFilter struct {
	Query      string     `json:"query"`
	MediaType  string     `json:"media_type"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	AlbumID    *uint      `json:"album_id"`
	Tags       []string   `json:"tags"`
	SortBy     string     `json:"sort_by"`
	SortOrder  string     `json:"sort_order"`
	PageSize   int        `json:"page_size"`
	PageNumber int        `json:"page_number"`
}

// SearchResult contains the search results
type SearchResult struct {
	Items     []models.MediaItem `json:"items"`
	Albums    []models.Album     `json:"albums"`
	Total     int64             `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
}

// NewSearchService creates a new search service instance
func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

// Search performs a search based on the provided filter
func (s *SearchService) Search(userID uint, filter SearchFilter) (*SearchResult, error) {
	var result SearchResult

	// Start building the query for media items
	query := s.db.Model(&models.MediaItem{}).Where("user_id = ?", userID)

	// Apply text search if query is provided
	if filter.Query != "" {
		query = query.Where(
			"filename LIKE ? OR metadata LIKE ?",
			"%"+filter.Query+"%",
			"%"+filter.Query+"%",
		)
	}

	// Apply media type filter
	if filter.MediaType != "" {
		query = query.Where("type = ?", filter.MediaType)
	}

	// Apply date range filter
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Apply album filter
	if filter.AlbumID != nil {
		query = query.Joins("JOIN album_media_items ON media_items.id = album_media_items.media_item_id").
			Where("album_media_items.album_id = ?", *filter.AlbumID)
	}

	// Apply tag filter
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("metadata LIKE ?", "%"+tag+"%")
		}
	}

	// Count total results before applying pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	result.Total = total

	// Apply sorting
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "DESC" {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, order))
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageNumber <= 0 {
		filter.PageNumber = 1
	}
	offset := (filter.PageNumber - 1) * filter.PageSize

	// Execute query with pagination
	var items []models.MediaItem
	if err := query.Offset(offset).Limit(filter.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	result.Items = items
	result.Page = filter.PageNumber
	result.PageSize = filter.PageSize

	// If searching by text, also search albums
	if filter.Query != "" {
		var albums []models.Album
		if err := s.db.Model(&models.Album{}).Where(
			"user_id = ? AND (name LIKE ? OR description LIKE ?)",
			userID,
			"%"+filter.Query+"%",
			"%"+filter.Query+"%",
		).Find(&albums).Error; err != nil {
			return nil, err
		}
		result.Albums = albums
	}

	return &result, nil
}

// SearchByMetadata searches for media items based on metadata fields
func (s *SearchService) SearchByMetadata(userID uint, metadata map[string]interface{}) ([]models.MediaItem, error) {
	var items []models.MediaItem
	query := s.db.Model(&models.MediaItem{}).Where("user_id = ?", userID)

	for key, value := range metadata {
		jsonQuery := fmt.Sprintf("JSON_EXTRACT(metadata, '$.%s') = ?", key)
		query = query.Where(jsonQuery, value)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// GetRecentMedia retrieves the most recently added media items
func (s *SearchService) GetRecentMedia(userID uint, limit int) ([]models.MediaItem, error) {
	var items []models.MediaItem
	if err := s.db.Model(&models.MediaItem{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetPopularTags returns the most frequently used tags
func (s *SearchService) GetPopularTags(userID uint) ([]string, error) {
	var items []models.MediaItem
	if err := s.db.Model(&models.MediaItem{}).
		Select("metadata").
		Where("user_id = ? AND metadata LIKE '%tags%'", userID).
		Find(&items).Error; err != nil {
		return nil, err
	}

	// Count tag frequencies
	tagFrequency := make(map[string]int)
	for _, item := range items {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
			continue
		}

		if tags, ok := metadata["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					tagFrequency[tagStr]++
				}
			}
		}
	}

	// Convert to slice and sort by frequency
	type tagCount struct {
		tag   string
		count int
	}
	var tagCounts []tagCount
	for tag, count := range tagFrequency {
		tagCounts = append(tagCounts, tagCount{tag: tag, count: count})
	}

	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].count > tagCounts[j].count
	})

	// Extract just the tags in order
	var tags []string
	for _, tc := range tagCounts {
		tags = append(tags, tc.tag)
	}

	// Return top 20 tags or all if less than 20
	if len(tags) > 20 {
		tags = tags[:20]
	}

	return tags, nil
}

// ParseMetadata extracts metadata from a JSON string
func ParseMetadata(metadataStr string) (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// BuildMetadata creates a JSON string from metadata fields
func BuildMetadata(fields map[string]interface{}) (string, error) {
	metadata, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(metadata), nil
}