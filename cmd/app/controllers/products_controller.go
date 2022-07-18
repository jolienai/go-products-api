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

func (controller ProductsController) GetProductBySku(c *gin.Context) {
	sku := c.Param("sku")

	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "sku is required",
		})
		return
	}

	products, err := controller.repository.GetProductBySkuQuery(sku)
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
