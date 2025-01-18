package response

type ValidationResponse struct {
	IsValid bool   `json:"IsValid"`
	Token   string `json:"Token"`
}
