package jobs

import (
	"fmt"
	"testing"

	"github.com/jolienai/go-products-api/cmd/app/dtos"
)

func TestDeduplicateProducts(t *testing.T) {
	t.Run("can add new guest", func(t *testing.T) {

		sample := []*dtos.ProductCsv{}

		sample = append(sample, &dtos.ProductCsv{Sku: "123", Name: "test", Country: "US", Quantity: 1})
		sample = append(sample, &dtos.ProductCsv{Sku: "123", Name: "test", Country: "US", Quantity: 1})

		fmt.Println(len(sample))

		result := deduplicateProducts(sample)

		if len(result) <= 0 {
			t.Errorf("Expected number greater than zero but got %d", len(result))
		}

		if (len(result)) > 1 {
			t.Errorf("Expected number greater than 1 but got %d", len(result))
		}

		if result[0].Quantity != 2 {
			t.Errorf("Quantity expected to be 2 but got %d", result[0].Quantity)
		}
	})
}
