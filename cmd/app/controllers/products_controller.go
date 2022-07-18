package controllers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"

	"github.com/jolienai/go-products-api/cmd/app/database"
	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

type ProductsController struct {
	repository database.ProductsRepository
}

func NewProductsController(repository database.ProductsRepository) *ProductsController {
	return &ProductsController{
		repository: repository,
	}
}

func (controller ProductsController) CosumeProduct(c *gin.Context) {
	request := dtos.CosumeProductRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		log.Fatal(err)
	}

	if request.Sku == "" || request.Quantity <= 0 || request.Country == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "sku, quantity and country are required",
		})
		return
	}

	getchannel := make(chan database.Product)
	go controller.repository.GetProductBySkuAndCountry(request.Sku, request.Country, getchannel)
	product := <-getchannel

	if request.Quantity > product.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product not found or stock insufficient",
		})
		return
	}

	updatechannel := make(chan bool)
	go controller.repository.UpdateProduct(product.Sku, product.Country, (product.Quantity - request.Quantity), updatechannel)
	updated := <-updatechannel

	c.JSON(http.StatusOK, gin.H{"result": updated})
}

func (controller ProductsController) GetProductBySku(c *gin.Context) {
	sku := c.Param("sku")

	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "sku is required",
		})
		return
	}

	products, err := controller.repository.GetProductBySku(sku)
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

func (controller ProductsController) UploadFile(c *gin.Context) {
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
			log.Println(err)
		}
	}

	fullpath := path + "/" + filename

	out, err := os.Create(fullpath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(out, file)
	if err != nil {
		log.Fatal(err)
	}
	out.Close()

	in, err := os.Open(fullpath)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	products := []*dtos.ProductCsv{}
	if err := gocsv.UnmarshalFile(in, &products); err != nil {
		panic(err)
	}

	controller.repository.AddFileToProccess(fullpath)

	c.JSON(http.StatusAccepted, gin.H{
		"message": fmt.Sprintf("File received: %s with %d products and will be processed soon :-)", filename, len(products)),
	})
}
