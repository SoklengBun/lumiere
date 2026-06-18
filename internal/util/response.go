package util

// Response is the standard API response envelope used across the project.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	CodeSuccess = 0
	CodeFailed  = -1

	// Common negative codes for project-level rules
	CodeBadRequest = -2
	CodeNotFound   = -3
	// CodeUnauthorized indicates the request is unauthenticated (user not logged in)
	CodeUnauthorized = -5
	CodeInternal     = -4
)

// DefaultMessages maps codes to default human-readable messages.
var DefaultMessages = map[int]string{
	CodeSuccess:      "success",
	CodeFailed:       "failed",
	CodeBadRequest:   "bad request",
	CodeUnauthorized: "user not login",
	CodeNotFound:     "not found",
	CodeInternal:     "internal server error",
}

// MessageForCode returns the provided custom message if non-empty, otherwise
// returns the default message for the code (or a generic "error").
func MessageForCode(code int, custom string) string {
	if custom != "" {
		return custom
	}
	if m, ok := DefaultMessages[code]; ok {
		return m
	}
	return "error"
}
