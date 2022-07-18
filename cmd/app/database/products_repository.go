package database

import (
	"github.com/jinzhu/gorm"

	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

type Product struct {
	gorm.Model
	Sku      string `gorm:"index"`
	Country  string
	Name     string
	Quantity int
}

type ProductBulkUpdateFromCsvFile struct {
	gorm.Model
	Filename string `gorm:"index"`
	Status   string
}

type ProductsRepository interface {
	GetProductBySkuQuery(sku string) (product []*Product, err error)
	AddFileToProccess(filepath string) error
	GetPedingCsvFile() (ProductBulkUpdateFromCsvFile, int64)
	UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool
	BulkProducts(productToUpdate []*dtos.ProductCsv) error
}

type database struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *database {
	return &database{
		db: db,
	}
}

func (database *database) AutoMigrate() {
	database.db.AutoMigrate(&Product{})
	database.db.AutoMigrate(&ProductBulkUpdateFromCsvFile{})
}

func (database *database) GetProductBySkuQuery(sku string) (product []*Product, err error) {
	var products []*Product
	if err := database.db.Where("sku = ?", sku).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (database *database) AddFileToProccess(filepath string) error {
	pendingCsvProductsFile := &ProductBulkUpdateFromCsvFile{Filename: filepath, Status: "pending"}
	database.db.Create(pendingCsvProductsFile)

	return nil
}

func (database *database) BulkProducts(productToUpdate []*dtos.ProductCsv) error {
	for _, p := range productToUpdate {
		product := &Product{Sku: p.Sku, Country: p.Country, Name: p.Name, Quantity: p.Quantity}
		if database.db.Model(&product).Where("sku = ? AND country = ?", p.Sku, p.Country).Updates(Product{Quantity: p.Quantity}).RowsAffected == 0 {
			database.db.Create(&product)
		}
	}

	return nil
}

func (database *database) GetPedingCsvFile() (ProductBulkUpdateFromCsvFile, int64) {
	var productToCreateOrUpdate = ProductBulkUpdateFromCsvFile{}
	result := database.db.Where("status=?", "pending").First(&productToCreateOrUpdate)

	return productToCreateOrUpdate, result.RowsAffected
}

func (database *database) UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool {
	if database.db.Model(&file).Where("id = ?", file.ID).Updates(ProductBulkUpdateFromCsvFile{Status: "processed"}).RowsAffected > 0 {
		return true
	}

	return false
}
