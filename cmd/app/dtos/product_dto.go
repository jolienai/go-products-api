package dtos

type ProductDto struct {
	Sku     string `json:"sku"`
	Country string `json:"country"`
	Name    string `json:"name"`
	Stock   int    `json:"quantity"`
}
