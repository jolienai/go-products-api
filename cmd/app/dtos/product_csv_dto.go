package dtos

type ProductCsv struct {
	Sku      string `csv:"sku"`
	Country  string `csv:"country"`
	Name     string `csv:"name"`
	Quantity int    `csv:"stock_change"`
}
