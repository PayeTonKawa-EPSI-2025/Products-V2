package integration_test

import (
	"context"
	"testing"

	"github.com/PayeTonKawa-EPSI-2025/Products-V2/internal/operation"
)

func TestIntegration_GetProducts(t *testing.T) {
	db := ConnectDB(t)
	ResetProductsTable(t, db)
	SeedDB(t, db)

	resp, err := operation.GetProducts(context.Background(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Body.Products) != 2 {
		t.Fatalf("expected 2 customers, got %d", len(resp.Body.Products))
	}
}
