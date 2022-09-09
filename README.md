# go-products-api

A simple example of how to implement a products API in Golang using:

- webframework
  - https://github.com/gin-gonic/gin
- Database
  - Postgres
- ORM
  - https://gorm.io/
  - Upsert using the gorm package v2, see more details here https://gorm.io/docs/v2_release_note.html#Upsert
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

3. PATCH /products
   Consume products if the is enough quantity in stock. It returns bad requests if the quantity is too high or product was not found in teh database.

Request body example:

```json
{
  "sku": "cbf87a9be799",
  "country": "dz",
  "quantity": 1
}
```

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

Note: You also can use the Postman collection products-api-postman_colq.json where you can find also an example of the PATCH method to consume a product and update its stock if possible.
