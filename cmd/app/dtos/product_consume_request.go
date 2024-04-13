package dtos

type ConsumeProductRequest struct {
	Country  string `json:"country"`
	Quantity int    `json:"quantity"`
}
