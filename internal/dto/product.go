package dto

import "github.com/PayeTonKawa-EPSI-2025/Common-V2/models"

type ProductsOutput struct {
	Body struct {
		Products []models.Product `json:"products"`
	}
}

type ProductOutput struct {
	Body models.Product
}

type ProductCreateInput struct {
	Body struct {
		Name    string                `json:"name"`
		Stock   uint                  `json:"stock"`
		Details models.ProductDetails `json:"details,omitempty"`
	}
}

type OrderProductsInput struct {
	OrderID uint `json:"orderId" path:"orderId"`
}
