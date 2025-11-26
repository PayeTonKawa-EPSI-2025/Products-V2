package operation_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PayeTonKawa-EPSI-2025/Products-V2/internal/operation"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbMock,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	return gormDB, mock
}

func TestGetProducts(t *testing.T) {
	db, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"name",
		"stock",
		"details_price",
		"details_description",
		"details_color",
	}).
		AddRow(1, nil, nil, nil, "Product A", 10, 19.99, "High quality product A", "Red").
		AddRow(2, nil, nil, nil, "Product B", 5, 9.99, "Affordable product B", "Blue")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).WillReturnRows(rows)

	resp, err := operation.GetProducts(context.Background(), db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Body.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp.Body.Products))
	}

	if resp.Body.Products[0].ID != 1 {
		t.Errorf("expected first product '1', got '%d'", resp.Body.Products[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

func TestGetProductNotFound(t *testing.T) {
	db, mock := setupMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1`)).
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := operation.GetProduct(context.Background(), db, 1)
	if err == nil {
		t.Fatal("expected error for non-existent product")
	}
}
