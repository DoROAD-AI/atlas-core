// types/types.go
package types

// ErrorResponse represents a generic error response structure for the API.
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
