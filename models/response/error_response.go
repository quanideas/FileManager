package response

type ErrorResponse struct {
	ErrorCode int    `json:"errorCode"`
	Error     string `json:"error"`
}
