package service

import (
	"fmt"

	"github.com/eralove/eralove-backend/internal/infrastructure/directus"
)

// CMSService handles CMS operations through Directus
type CMSService struct {
	directusClient *directus.Client
}

// NewCMSService creates a new CMS service
func NewCMSService(directusClient *directus.Client) *CMSService {
	return &CMSService{
		directusClient: directusClient,
	}
}

// GetContent retrieves content from a collection
func (s *CMSService) GetContent(collection string, filters map[string]string) (interface{}, error) {
	return s.directusClient.GetItems(collection, filters)
}

// GetContentByID retrieves a single content item by ID
func (s *CMSService) GetContentByID(collection, id string) (interface{}, error) {
	return s.directusClient.GetItem(collection, id)
}

// CreateContent creates new content in a collection
func (s *CMSService) CreateContent(collection string, data interface{}) (interface{}, error) {
	return s.directusClient.CreateItem(collection, data)
}

// UpdateContent updates existing content
func (s *CMSService) UpdateContent(collection, id string, data interface{}) (interface{}, error) {
	return s.directusClient.UpdateItem(collection, id, data)
}

// DeleteContent deletes content from a collection
func (s *CMSService) DeleteContent(collection, id string) error {
	return s.directusClient.DeleteItem(collection, id)
}

// GetFiles retrieves files from Directus
func (s *CMSService) GetFiles(filters map[string]string) (interface{}, error) {
	return s.directusClient.GetFiles(filters)
}

// Example: Get blog posts
func (s *CMSService) GetBlogPosts(limit, offset int, status string) (interface{}, error) {
	filters := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}
	if status != "" {
		filters["filter[status][_eq]"] = status
	}
	return s.directusClient.GetItems("posts", filters)
}

// Example: Get pages
func (s *CMSService) GetPages() (interface{}, error) {
	return s.directusClient.GetItems("pages", nil)
}

// Example: Get settings
func (s *CMSService) GetSettings() (interface{}, error) {
	return s.directusClient.GetItems("settings", nil)
}
