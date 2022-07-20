package database

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

type Product struct {
	gorm.Model
	Sku      string `gorm:"index:,unique,composite:myname"`
	Country  string `gorm:"index:,unique,composite:myname"`
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
	AddFileToProccess(filepath string) error
	GetPedingCsvFile() (ProductBulkUpdateFromCsvFile, int64)
	UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool
	BulkProducts(productToUpdate []*dtos.ProductCsv) error
	UpsertProducts(productToUpdate []*dtos.ProductCsv) error
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

func (database *database) UpdateProduct(sku string, country string, quantity int, updated chan<- bool) {
	product := Product{Sku: sku, Country: country, Quantity: quantity}
	if database.db.Model(&product).Where("sku = ? AND country = ?", sku, country).Update("quantity", quantity).RowsAffected == 0 {
		updated <- false
	}
	updated <- true
}

func (database *database) GetProductBySkuAndCountry(sku string, country string, product chan<- Product) {
	productFromDb := Product{}
	result := database.db.Where("sku = ? AND country = ?", sku, country).First(&productFromDb)
	if result.RowsAffected == 0 {
		product <- Product{Quantity: 0}
	}
	product <- productFromDb
}

func (database *database) GetProductBySku(sku string) (product []*Product, err error) {
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

func (database *database) UpsertProducts(productToUpdate []*dtos.ProductCsv) error {
	products := make([]Product, 0)
	for _, p := range productToUpdate {
		product := Product{Sku: p.Sku, Country: p.Country, Name: p.Name, Quantity: p.Quantity}
		products = append(products, product)
	}

	//TODO: getting ERROR: ON CONFLICT DO UPDATE command cannot affect row a second time (SQLSTATE 21000)
	// when processing the secound file
	// issue related https://github.com/go-gorm/gorm/issues/5007
	err := database.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku"}, {Name: "country"}},
		DoUpdates: clause.AssignmentColumns([]string{"quantity"}),
	}).CreateInBatches(&products, 5000).Error

	if err != nil {
		return err
	}

	return nil
}

func (database *database) GetPedingCsvFile() (ProductBulkUpdateFromCsvFile, int64) {
	var productToCreateOrUpdate = ProductBulkUpdateFromCsvFile{}
	result := database.db.Where("status=?", "pending").First(&productToCreateOrUpdate)

	return productToCreateOrUpdate, result.RowsAffected
}

func (database *database) UpdateCsvFileStatusToProcessed(file ProductBulkUpdateFromCsvFile) bool {
	return database.db.Model(&file).Where("id = ?", file.ID).Updates(ProductBulkUpdateFromCsvFile{Status: "processed"}).RowsAffected > 0
}
