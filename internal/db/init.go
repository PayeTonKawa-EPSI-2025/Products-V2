package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/PayeTonKawa-EPSI-2025/Common-V2/models"

	localModels "github.com/PayeTonKawa-EPSI-2025/Products-V2/internal/models"
)

func Init() *gorm.DB {
	dsn := os.Getenv("DATABASE_DSN")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	db.AutoMigrate(&models.Product{}, &localModels.Customer{}, &localModels.Order{}, &localModels.OrderProduct{})

	return db
}
