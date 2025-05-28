package models

// HTTPError represents an HTTP error response
// swagger:model HTTPError
type HTTPError struct {
	// HTTP status code
	Code int `json:"code"`
	// Error message
	Message string `json:"message"`
}