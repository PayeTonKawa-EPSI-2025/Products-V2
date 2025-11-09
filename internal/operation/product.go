package operation

import (
	"context"
	"errors"
	"net/http"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	"github.com/PayeTonKawa-EPSI-2025/Products/internal/dto"
	localModels "github.com/PayeTonKawa-EPSI-2025/Products/internal/models"
	"github.com/PayeTonKawa-EPSI-2025/Products/internal/rabbitmq"
	"github.com/danielgtaylor/huma/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// ----------------------
// Extracted CRUD Functions
// ----------------------

// Get all orders
func GetProducts(ctx context.Context, db *gorm.DB) (*dto.ProductsOutput, error) {
	resp := &dto.ProductsOutput{}

	var products []models.Product
	results := db.Find(&products)

	if results.Error == nil {
		resp.Body.Products = products
	}

	return resp, results.Error
}

// Get a single product by ID
func GetProduct(ctx context.Context, db *gorm.DB, id uint) (*dto.ProductOutput, error) {
	resp := &dto.ProductOutput{}

	var product models.Product
	results := dbConn.First(&product, id)

	if results.Error == nil {
		resp.Body = product
		return resp, nil
	}

	if errors.Is(results.Error, gorm.ErrRecordNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "Product not found")
	}

	return nil, results.Error
}

func GetProductsByIdOrder(ctx context.Context, db *gorm.DB, id uint) (*dto.ProductsOutput, error) {
	resp := &dto.ProductsOutput{}

	var orderProducts []localModels.OrderProduct
	if err := db.Where("order_id = ?", id).Find(&orderProducts).Error; err != nil {
		return nil, err
	}

	productIDs := make([]uint, 0, len(orderProducts))
	for _, op := range orderProducts {
		productIDs = append(productIDs, op.ProductID)
	}

	var products []models.Product
	if len(productIDs) > 0 {
		if err := db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			return nil, err
		}
	}

	resp.Body.Products = products

	return resp, nil
}

// ----------------------
// Register routes with Huma
// ----------------------

func RegisterProductsRoutes(api huma.API, dbConn *gorm.DB, ch *amqp.Channel) {

	huma.Register(api, huma.Operation{
		OperationID: "get-products",
		Summary:     "Get all products",
		Method:      http.MethodGet,
		Path:        "/products",
		Tags:        []string{"products"},
	}, func(ctx context.Context, input *struct{}) (*dto.ProductsOutput, error) {
		return GetProducts(ctx, dbConn)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-product",
		Summary:     "Get a product",
		Method:      http.MethodGet,
		Path:        "/products/{id}",
		Tags:        []string{"products"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*dto.ProductOutput, error) {
		return GetProduct(ctx, dbConn, input.Id)
	})

	huma.Register(api, huma.Operation{
		OperationID:   "get-orders-products",
		Summary:       "Get all products for an order",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
		Path:          "/products/{orderId}/orders",
		Tags:          []string{"products"},
	}, func(ctx context.Context, input *dto.OrderProductsInput) (*dto.ProductsOutput, error) {
		return GetProductsByIdOrder(ctx, dbConn, input.OrderID)
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-product",
		Summary:       "Create a product",
		Method:        http.MethodPost,
		DefaultStatus: http.StatusCreated,
		Path:          "/products",
		Tags:          []string{"products"},
	}, func(ctx context.Context, input *dto.ProductCreateInput) (*dto.ProductOutput, error) {
		resp := &dto.ProductOutput{}

		product := models.Product{
			Name:    input.Body.Name,
			Stock:   input.Body.Stock,
			Details: input.Body.Details,
		}

		results := dbConn.Create(&product)

		if results.Error == nil {
			resp.Body = product

			// Publish product created event
			err := rabbitmq.PublishProductEvent(ch, events.ProductCreated, product)
			if err != nil {
				// Log the error but don't fail the request
				// The product was already created in the database
				return resp, nil
			}
		}

		return resp, results.Error
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-product",
		Summary:     "Replace a product",
		Method:      http.MethodPut,
		Path:        "/products/{id}",
		Tags:        []string{"products"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
		dto.ProductCreateInput
	}) (*dto.ProductOutput, error) {
		resp := &dto.ProductOutput{}

		var product models.Product
		results := dbConn.First(&product, input.Id)

		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Product not found")
		}
		if results.Error != nil {
			return nil, results.Error
		}

		updates := models.Product{
			Name:    input.Body.Name,
			Stock:   input.Body.Stock,
			Details: input.Body.Details,
		}

		results = dbConn.Model(&product).Updates(updates)
		if results.Error != nil {
			return nil, results.Error
		}

		// Get updated product from DB to ensure all fields are correct
		dbConn.First(&product, product.ID)
		resp.Body = product

		// Publish product updated event
		err := rabbitmq.PublishProductEvent(ch, events.ProductUpdated, product)
		if err != nil {
			// Log the error but don't fail the request
			// The product was already updated in the database
		}

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-product",
		Summary:       "Delete a product",
		Method:        http.MethodDelete,
		DefaultStatus: http.StatusNoContent,
		Path:          "/products/{id}",
		Tags:          []string{"products"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*struct{}, error) {
		resp := &struct{}{}

		// First get the product to have the complete data for the event
		var product models.Product
		result := dbConn.First(&product, input.Id)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Product not found")
		}

		if result.Error != nil {
			return nil, result.Error
		}

		results := dbConn.Delete(&product)

		if results.Error == nil {
			// Publish product deleted event
			err := rabbitmq.PublishProductEvent(ch, events.ProductDeleted, product)
			if err != nil {
				// Log the error but don't fail the request
				// The product was already deleted from the database
			}

			return resp, nil
		}

		return nil, results.Error
	})
}
