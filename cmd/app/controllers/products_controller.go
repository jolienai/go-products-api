package controllers

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"

	"github.com/jolienai/go-products-api/cmd/app/database"
	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

type ProductsController struct {
	repository database.ProductsRepository
	logger     zap.Logger
}

func NewProductsController(repository database.ProductsRepository, logger *zap.Logger) *ProductsController {
	return &ProductsController{
		repository: repository,
		logger:     *logger,
	}
}

func (p ProductsController) ConsumeProduct(c *gin.Context) {
	sku := c.Param("sku")
	request := dtos.ConsumeProductRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		p.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "not possible to parse request",
		})
		return
	}

	if sku == "" || request.Quantity <= 0 || request.Country == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "sku, quantity and country are required",
		})
		return
	}

	get := make(chan database.Product)
	go p.repository.GetProductBySkuAndCountry(sku, request.Country, get)
	product := <-get

	if request.Quantity > product.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product not found or stock insufficient",
		})
		return
	}

	update := make(chan bool)
	go p.repository.UpdateProduct(product.Sku, product.Country, product.Quantity-request.Quantity, update)
	productUpdated := <-update

	c.JSON(http.StatusOK, gin.H{"result": productUpdated})
}

func (p ProductsController) GetProductBySku(c *gin.Context) {
	sku := c.Param("sku")

	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "sku is required",
		})
		return
	}

	products, err := p.repository.GetProductBySku(sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	productsDto := make([]dtos.ProductDto, 0, len(products))
	for _, p := range products {
		productsDto = append(productsDto, dtos.ProductDto{Sku: p.Sku, Name: p.Name, Country: p.Country, Stock: p.Quantity})
	}

	c.JSON(http.StatusOK, gin.H{
		"products": productsDto,
	})
}

func (p ProductsController) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("file err : %s", err.Error()))
		return
	}
	filename := header.Filename

	path := "product_bulk_files"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModeDir)
		if err != nil {
			p.logger.Error(err.Error())
			c.String(http.StatusInternalServerError, fmt.Sprintf("Mkdir err : %s", err.Error()))
			return
		}
	}

	filepath := path + "/" + filename

	fileTobeProcessed, err := os.Create(filepath)
	if err != nil {
		p.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("Create file err : %s", err.Error()))
		return
	}

	_, err = io.Copy(fileTobeProcessed, file)
	if err != nil {
		p.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("Copy file err : %s", err.Error()))
		return
	}
	err = fileTobeProcessed.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Close file err : %s", err.Error()))
		return
	}

	fileSaved, err := os.Open(filepath)
	if err != nil {
		p.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("Open saved file err : %s", err.Error()))
		return
	}
	defer func(fileSaved *os.File) {
		err := fileSaved.Close()
		if err != nil {
			p.logger.Error(err.Error())
			c.String(http.StatusInternalServerError, fmt.Sprintf("Close saved file err : %s", err.Error()))
			return
		}
	}(fileSaved)

	var products []*dtos.ProductCsv
	if err := gocsv.UnmarshalFile(fileSaved, &products); err != nil {
		panic(err)
	}

	err = p.repository.AddFileToProcess(filepath)
	if err != nil {
		c.JSON(http.StatusAccepted, gin.H{
			"message": fmt.Sprintf("Error when saving the file: %s with %d records :-(", filename, len(products)),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": fmt.Sprintf("File received: %s with %d products and will be processed soon :-)", filename, len(products)),
	})
}
