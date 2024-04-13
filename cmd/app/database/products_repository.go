package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

type Product struct {
	gorm.Model
	Sku      string `gorm:"index:,unique,composite:sku_contry"`
	Country  string `gorm:"index:,unique,composite:sku_contry"`
	Name     string
	Quantity int
}

type ProductBulkUpdateFromCsvFile struct {
	gorm.Model
	Filename string `gorm:"index"`
	Status   string
}

type ProductsRepository interface {
	GetProductBySkuAndCountry(sku string, country string, product chan<- Product)
	GetProductBySku(sku string) (product []*Product, err error)
	UpdateProduct(sku string, country string, quantity int, done chan<- bool)
	AddFileToProcess(filepath string) error
	GetPendingCsvFile() (ProductBulkUpdateFromCsvFile, int64)
	UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool
	UpsertProducts(productToUpdate []*dtos.ProductCsv) error
}

type repository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) AutoMigrate() error {
	err := repo.db.AutoMigrate(&Product{})
	if err != nil {
		return err
	}
	err = repo.db.AutoMigrate(&ProductBulkUpdateFromCsvFile{})
	if err != nil {
		return err
	}
	return nil
}

func (repo *repository) UpdateProduct(sku string, country string, quantity int, updated chan<- bool) {
	product := Product{Sku: sku, Country: country, Quantity: quantity}
	if repo.db.Model(&product).Where("sku = ? AND country = ?", sku, country).Update("quantity", quantity).RowsAffected == 0 {
		updated <- false
	}
	updated <- true
}

func (repo *repository) GetProductBySkuAndCountry(sku string, country string, product chan<- Product) {
	productFromDb := Product{}
	result := repo.db.Where("sku = ? AND country = ?", sku, country).First(&productFromDb)
	if result.RowsAffected == 0 {
		product <- Product{Quantity: 0}
	}
	product <- productFromDb
}

func (repo *repository) GetProductBySku(sku string) (product []*Product, err error) {
	var products []*Product
	if err := repo.db.Where("sku = ?", sku).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (repo *repository) AddFileToProcess(filepath string) error {
	pendingCsvProductsFile := &ProductBulkUpdateFromCsvFile{Filename: filepath, Status: "pending"}
	repo.db.Create(pendingCsvProductsFile)

	return nil
}

func (repo *repository) UpsertProducts(productToUpdate []*dtos.ProductCsv) error {
	products := make([]Product, 0)
	for _, p := range productToUpdate {
		products = append(products, Product{Sku: p.Sku, Country: p.Country, Name: p.Name, Quantity: p.Quantity})
	}

	err := repo.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku"}, {Name: "country"}},
		DoUpdates: clause.AssignmentColumns([]string{"quantity"}),
	}).CreateInBatches(&products, 5000).Error

	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) GetPendingCsvFile() (ProductBulkUpdateFromCsvFile, int64) {
	var productToCreateOrUpdate = ProductBulkUpdateFromCsvFile{}
	result := repo.db.Where("status=?", "pending").First(&productToCreateOrUpdate)

	return productToCreateOrUpdate, result.RowsAffected
}

func (repo *repository) UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool {
	return repo.db.Model(&file).Where("id = ?", file.ID).Updates(ProductBulkUpdateFromCsvFile{Status: "processed"}).RowsAffected > 0
}
