package jobs

import (
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"go.uber.org/zap"

	"github.com/jolienai/go-products-api/cmd/app/database"
	"github.com/jolienai/go-products-api/cmd/app/dtos"

	"github.com/go-co-op/gocron"
)

type Job struct {
	repository database.ProductsRepository
	logger     zap.Logger
}

func NewJob(repository database.ProductsRepository, logger *zap.Logger) *Job {
	return &Job{
		repository: repository,
		logger:     *logger,
	}
}

func (job Job) ProductBulkJob() error {
	csv, rows := job.repository.GetPedingCsvFile()
	if rows > 0 {
		job.logger.Info(fmt.Sprintf("found file pending: %s and ID: %d", csv.Filename, csv.ID))

		file, err := os.Open(csv.Filename)
		if err != nil {
			return err
		}
		defer file.Close()

		products := []*dtos.ProductCsv{}
		if err := gocsv.UnmarshalFile(file, &products); err != nil {
			return err
		}

		unique := deduplicateProducts(products)

		job.logger.Info(fmt.Sprintf("Processing: %s with %d rows", csv.Filename, len(products)))
		err = job.repository.UpsertProducts(unique)
		if err != nil {
			return err
		}

		if job.repository.UpdateCsvFileStatusToProcessed(csv) {
			job.logger.Info(fmt.Sprintf("File processed: %s", csv.Filename))
		}
	}
	return nil
}

func deduplicateProducts(products []*dtos.ProductCsv) []*dtos.ProductCsv {
	var unique []*dtos.ProductCsv
	type key struct{ value1, value2 string }
	m := make(map[key]int)
	for _, product := range products {
		k := key{product.Sku, product.Country}
		if i, ok := m[k]; ok {
			unique[i].Quantity = unique[i].Quantity + product.Quantity
		} else {
			m[k] = len(unique)
			unique = append(unique, product)
		}
	}
	return unique
}

func (job Job) ScheduleProductBulkJob() error {
	scheduler := gocron.NewScheduler(time.UTC).SingletonMode()
	//TODO: add the time in some configurations
	_, err := scheduler.Every("1m").Do(func() {
		joberr := job.ProductBulkJob()
		if joberr != nil {
			fmt.Println(joberr.Error())
		}
	})
	if err != nil {
		return err
	}
	scheduler.StartAsync()
	return nil
}
