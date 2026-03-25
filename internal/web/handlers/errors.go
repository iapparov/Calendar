package handlers

// errorResponse is a typed struct for JSON error replies.
// Using a struct instead of gin.H (map[string]any) avoids 1 map allocation per error response.
type errorResponse struct {
	Error string `json:"error"`
}

// Pre-allocated static error responses — zero alloc on the hot path.
var (
	errUserIDNotFound = &errorResponse{Error: "user_id not found in context"}
	errDateRequired   = &errorResponse{Error: "date query parameter is required"}
)
