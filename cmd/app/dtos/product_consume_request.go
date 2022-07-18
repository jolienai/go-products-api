package dtos

type CosumeProductRequest struct {
	Sku      string `json:"sku"`
	Country  string `json:"country"`
	Quantity int    `json:"quantity"`
}
