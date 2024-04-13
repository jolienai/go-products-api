package main

import (
	"github.com/joho/godotenv"
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

	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	logger.Info("Starting products-api...")

	dsn := os.Getenv("POSTGRES_CONNECTION_STRING")
	if dsn == "" {
		logger.Fatal("Postgres connection string is not set")
		panic("POSTGRES_CONNECTION_STRING environment variable must be set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		panic(err.Error())
	}

	logger.Info("Trying to apply the migrations..." + dsn)
	repository := database.NewProductsRepository(db)
	err = repository.AutoMigrate()
	if err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
		panic(err.Error())
	}
	logger.Info("Migrations applied...")

	logger.Info("Starting jobs...")
	job := jobs.NewJob(repository, logger)
	err = job.ScheduleProductBulkJob()
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info("Jobs are ready...")

	controller := controllers.NewProductsController(repository, logger)
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.POST("/products/bulk", controller.UploadFile)
		v1.GET("/products/:sku", controller.GetProductBySku)
		v1.PATCH("/products/:sku/consume", controller.ConsumeProduct)
	}
	err = router.Run(":8080")
	if err != nil {
		logger.Fatal("error trying to start the app", zap.Error(err))
		return
	}
}
