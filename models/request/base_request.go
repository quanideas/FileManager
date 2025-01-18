package request

import "github.com/google/uuid"

type GetByIDRequest struct {
	ID string
}

type DeleteByIDRequest struct {
	ID uuid.UUID `json:"id"`
}

type GetAll struct {
	Page  int
	Count int
	Sort  []struct {
		By   string
		Type string
	}
	Search []struct {
		By    string
		Value string
	}
	CompanyID string `json:"company_id"`
}
