package operation

import "github.com/PayeTonKawa-EPSI-2025/Common/models"

// Mock function to replace PublishCustomerEvent
var PublishProductEvent = func(ch any, event string, customer models.Product) error {
	return nil
}
