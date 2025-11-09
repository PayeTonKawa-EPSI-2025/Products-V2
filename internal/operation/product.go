package operation

import (
	"context"
	"errors"
	"net/http"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	"github.com/PayeTonKawa-EPSI-2025/Products/internal/dto"
	"github.com/PayeTonKawa-EPSI-2025/Products/internal/rabbitmq"
	"github.com/danielgtaylor/huma/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

func RegisterProductsRoutes(api huma.API, dbConn *gorm.DB, ch *amqp.Channel) {

	huma.Register(api, huma.Operation{
		OperationID: "get-products",
		Summary:     "Get all products",
		Method:      http.MethodGet,
		Path:        "/products",
		Tags:        []string{"products"},
	}, func(ctx context.Context, input *struct{}) (*dto.ProductsOutput, error) {
		resp := &dto.ProductsOutput{}

		var products []models.Product
		results := dbConn.Find(&products)

		if results.Error == nil {
			resp.Body.Products = products
		}

		return resp, results.Error
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
		resp := &dto.ProductOutput{}

		var product models.Product
		results := dbConn.First(&product, input.Id)

		if results.Error == nil {
			resp.Body = product
			return resp, nil
		}

		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Product not found")
		}

		return nil, results.Error
	})

	huma.Register(api, huma.Operation{
		OperationID:   "get-orders-products",
		Summary:       "Get all products for an order",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
		Path:          "/products/{orderId}/orders",
		Tags:          []string{"products"},
	}, func(ctx context.Context, input *dto.OrderProductsInput) (*dto.ProductsOutput, error) {
		resp := &dto.ProductsOutput{}

		var products []models.Product
		if err := dbConn.Where("order_id = ?", input.OrderID).Find(&products).Error; err != nil {
			return nil, err
		}

		resp.Body.Products = products

		return resp, nil
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
