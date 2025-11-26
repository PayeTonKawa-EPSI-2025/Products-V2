package rabbitmq

import (
	"github.com/PayeTonKawa-EPSI-2025/Products-V2/internal/rabbitmq/event_handlers"
	"gorm.io/gorm"
)

// SetupEventHandlers configures handlers for different event types
func SetupEventHandlers(dbConn *gorm.DB) *EventRouter {
	router := NewEventRouter()

	// Initialize event handlers
	customerHandlers := event_handlers.NewCustomerEventHandlers(dbConn)
	orderHandlers := event_handlers.NewOrderEventHandlers(dbConn)
	debugHandlers := event_handlers.NewDebugEventHandlers()

	// Register customer event handlers
	router.RegisterHandler("customer.created", customerHandlers.HandleCustomerCreated)
	router.RegisterHandler("customer.updated", customerHandlers.HandleCustomerUpdated)
	router.RegisterHandler("customer.deleted", customerHandlers.HandleCustomerDeleted)

	// Register order event handlers
	router.RegisterHandler("order.created", orderHandlers.HandleOrderCreated)
	router.RegisterHandler("order.updated", orderHandlers.HandleOrderUpdated)
	router.RegisterHandler("order.deleted", orderHandlers.HandleOrderDeleted)

	// Register debug catch-all handler
	// Useful during development, can be removed in production
	router.RegisterHandler("#", debugHandlers.HandleAllEvents)

	return router
}
