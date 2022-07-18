# go-products-api

A simple example of how to implement a products API in Golang using:

- webframework
  - https://github.com/gin-gonic/gin
- Database
  - Postgres
- ORM
  - https://gorm.io/
- Logger

  - https://pkg.go.dev/go.uber.org/zap@v1.21.0
  - https://pkg.go.dev/go.elastic.co/ecszap@v1.0.1

- Jobs
  - https://pkg.go.dev/github.com/go-co-op/gocron@v1.15.0
  - https://github.com/go-co-op/gocron
- Docker and docker-compose

## Endpoints

1. POST /products/bulk
   Using to upload the csv file to be processed, so the stock for each product will be updated accordingly.

the .csv file should contain the following structure:

```csv
country,sku,name,stock_change
"gh","e920c573f128","Ramirez-Molina Granite Pizza","100"
```

2. GET /products/:sku
   Returns a list of products for the specified sku, passed in the url

## Run products-api locally

1. Set environment variable

```bash
export POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=testuser dbname=productsdb password=123456 sslmode=disable"
```

```bash
make postgres-up
```

```bash
cd cmd/app
go run main.go
```

## Testing

got to `http://localhost:8080/v1/products/bulk` to upload the csv file, after that the job will update the product accordingly.
got to `http://localhost:8080/v1/products/{put_here_product_sku}` to find a list of products for the specified sku.

See an example below:

```bash
curl GET http://localhost:8080/v1/products/e920c573f128
```

```json
{
  "products": [
    {
      "sku": "e920c573f128",
      "country": "gh",
      "name": "Ramirez-Molina Granite Pizza",
      "stock_change": 51
    },
    {
      "sku": "e920c573f128",
      "country": "ma",
      "name": "Ramirez-Molina Granite Pizza",
      "stock_change": 58
    },
    {
      "sku": "e920c573f128",
      "country": "ug",
      "name": "Ramirez-Molina Granite Pizza",
      "stock_change": 63
    }
  ]
}
```

## Known issues

- The connection between the postgres container and the app, which is not working now.

### Improvements

- Improve bulk insert and update performance, even tough it is done 34k products in less than 2 min done by the job.
- The minutes set in the cron job scheduler should be inside a configuration file or env variable, for instance.

logs for the job

```json
{"log.level":"info","@timestamp":"2022-07-18T19:39:51.977+0100","log.origin":{"file.name":"jobs/product_bulk_job.go","file.line":45},"message":"Processing: product_bulk_files/challenge_1_ecommerce_stock_file_1.csv with 31995 rows","app":"products-api","environment":"local","ecs.version":"1.6.0"}

{"log.level":"info","@timestamp":"2022-07-18T19:41:12.421+0100","log.origin":{"file.name":"jobs/product_bulk_job.go","file.line":52},"message":"File processed: product_bulk_files/challenge_1_ecommerce_stock_file_1.csv","app":"products-api","environment":"local","ecs.version":"1.6.0"}
```
