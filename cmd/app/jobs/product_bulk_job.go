package jobs

import (
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"

	"github.com/jolienai/go-products-api/cmd/app/database"
	"github.com/jolienai/go-products-api/cmd/app/dtos"

	"github.com/go-co-op/gocron"
)

type Job struct {
	repository database.ProductsRepository
}

func NewJob(repository database.ProductsRepository) *Job {
	return &Job{
		repository: repository,
	}
}

func (job Job) ProductBulkJob() error {

	csv, rows := job.repository.GetPedingCsvFile()

	if rows > 0 {

		fmt.Println(fmt.Sprintf("found file pending: %s and ID: %d", csv.Filename, csv.ID))

		file, err := os.Open(csv.Filename)
		if err != nil {
			return err
		}
		defer file.Close()

		products := []*dtos.ProductCsv{}
		if err := gocsv.UnmarshalFile(file, &products); err != nil {
			return err
		}

		fmt.Println(fmt.Sprintf("Processing: %s with %d rows", csv.Filename, len(products)))
		err = job.repository.BulkProducts(products)
		if err != nil {
			return err
		}

		// update status
		if job.repository.UpdateCsvFileStatusToProcessed(csv) {
			fmt.Println(fmt.Sprintf("File processed: %s", csv.Filename))
		}
	}
	return nil
}

func (job Job) ScheduleProductBulkJob() error {
	scheduler := gocron.NewScheduler(time.UTC).SingletonMode()
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
