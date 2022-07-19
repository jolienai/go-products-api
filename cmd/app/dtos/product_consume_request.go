package dtos

type CosumeProductRequest struct {
	Country  string `json:"country"`
	Quantity int    `json:"quantity"`
}
