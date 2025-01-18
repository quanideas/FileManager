package request

type ValidationRequest struct {
	Token        string `json:"Token"`
	RefreshToken string `json:"RefreshToken"`
}
