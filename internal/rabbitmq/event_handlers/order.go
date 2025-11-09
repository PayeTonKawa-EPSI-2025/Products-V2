package event_handlers

import (
	"encoding/json"
	"log"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	localModels "github.com/PayeTonKawa-EPSI-2025/Products/internal/models"
	"gorm.io/gorm"
)

// OrderEventHandlers provides handlers for order-related events
type OrderEventHandlers struct {
	db *gorm.DB
}

// NewOrderEventHandlers creates a new order event handlers instance
func NewOrderEventHandlers(db *gorm.DB) *OrderEventHandlers {
	return &OrderEventHandlers{db: db}
}

// HandleOrderCreated handles the order.created event
func (h *OrderEventHandlers) HandleOrderCreated(body []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling order.created event: %v", err)
		return err
	}

	log.Printf("Received order.created event for order %d", event.Order.OrderID)

	// Create the order in the local database
	order := localModels.Order{}
	order.ID = event.Order.OrderID

	if err := h.db.Create(&order).Error; err != nil {
		log.Printf("Error creating order in DB: %v", err)
		return err
	}

	log.Printf("Successfully created order %d in local database", order.ID)

	// Begin a transaction
	tx := h.db.Begin()

	if tx.Error != nil {
		log.Printf("Failed to start transaction: %v\n", tx.Error)
		return nil
	}

	var orderProducts []localModels.OrderProduct

	for _, productID := range event.Order.ProductIDs {
		// Decrement stock safely (only if stock > 0)
		result := tx.Model(&models.Product{}).
			Where("id = ? AND stock > 0", productID).
			UpdateColumn("stock", gorm.Expr("stock - ?", 1))

		if result.Error != nil {
			tx.Rollback()
			log.Printf("Failed to update stock for product %d: %v\n", productID, result.Error)
			return nil
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Printf("Product %d has no stock left\n", productID)
			return nil
		}

		orderProducts = append(orderProducts, localModels.OrderProduct{
			OrderID:   event.Order.OrderID,
			ProductID: productID,
		})
	}

	// Create all OrderProduct links
	if err := tx.Create(&orderProducts).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to create OrderProduct records: %v\n", err)
		return nil
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction: %v\n", err)
		return nil
	}

	log.Printf("Successfully created order products for order %d in local database", order.ID)

	return nil
}

// HandleOrderUpdated handles the order.updated event
func (h *OrderEventHandlers) HandleOrderUpdated(body []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling order.updated event: %v", err)
		return err
	}

	log.Printf("Received order.updated event for order %d", event.Order.OrderID)

	// Update the order in the local database
	order := localModels.Order{}
	order.ID = event.Order.OrderID

	if err := h.db.Save(&order).Error; err != nil {
		log.Printf("Error updating order in DB: %v", err)
		return err
	}

	log.Printf("Successfully updated order %d in local database", order.ID)
	return nil
}

// HandleOrderDeleted handles the order.deleted event
func (h *OrderEventHandlers) HandleOrderDeleted(body []byte) error {
	var event events.OrderEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling order.deleted event: %v", err)
		return err
	}

	log.Printf("Received order.deleted event for order %d", event.Order.OrderID)

	// Delete the order from the local database
	if err := h.db.Delete(&localModels.Order{}, event.Order.OrderID).Error; err != nil {
		log.Printf("Error deleting order from DB: %v", err)
		return err
	}

	log.Printf("Successfully deleted order %d from local database", event.Order.OrderID)
	return nil
}
