package main

import (
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"

	"github.com/jolienai/go-products-api/cmd/app/controllers"

	"github.com/jolienai/go-products-api/cmd/app/database"

	"github.com/jolienai/go-products-api/cmd/app/jobs"

	"github.com/jinzhu/gorm"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
)

func main() {

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	logger = logger.With(zap.String("app", "myapp")).With(zap.String("environment", "psm"))

	logger.Info("Starting products-api...")

	connectionString := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connectionString == "" {
		panic("POSTGRES_CONNECTION_STRING environment variable must be set")
	}

	logger.Error("connectionString")

	//db, err := gorm.Open("postgres", "host=localhost port=5432 user=testuser dbname=productsdb password=123456 sslmode=disable")
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		panic(err.Error())
	}

	logger.Info("Trying to apply the migrations..." + connectionString)
	repository := database.NewProductsRepository(db)
	repository.AutoMigrate()
	logger.Info("Migrations applied...")

	logger.Info("Starting jobs...")
	job := jobs.NewJob(repository)
	job.ScheduleProductBulkJob()
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info("Jobs started...")

	// controllers
	controller := controllers.NewProductsController(repository)

	// routes
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.POST("/products/bulk", controller.UploadFile)
		v1.GET("/products/:sku", controller.GetProductBySku)
	}

	router.Run(":8080")
}
