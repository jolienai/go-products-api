package main

import (
	"os"

	"github.com/gin-gonic/gin"

	"github.com/jolienai/go-products-api/cmd/app/controllers"
	"github.com/jolienai/go-products-api/cmd/app/jobs"

	"github.com/jolienai/go-products-api/cmd/app/database"

	"gorm.io/gorm"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"

	"gorm.io/driver/postgres"
)

func main() {
	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	logger = logger.With(zap.String("app", "products-api")).With(zap.String("environment", "local"))

	logger.Info("Starting products-api...")

	dsn := os.Getenv("POSTGRES_CONNECTION_STRING")
	if dsn == "" {
		panic("POSTGRES_CONNECTION_STRING environment variable must be set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	logger.Info("Trying to apply the migrations..." + dsn)
	repository := database.NewProductsRepository(db)
	repository.AutoMigrate()
	logger.Info("Migrations applied...")

	logger.Info("Starting jobs...")
	job := jobs.NewJob(repository, logger)
	job.ScheduleProductBulkJob()
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info("Jobs started...")

	controller := controllers.NewProductsController(repository)
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.POST("/products/bulk", controller.UploadFile)
		v1.GET("/products/:sku", controller.GetProductBySku)
		v1.PATCH("/products/:sku/consume", controller.CosumeProduct)
	}
	router.Run(":8080")
}
