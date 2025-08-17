package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"index:idx_code_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Code  string `gorm:"index:idx_code_updated"`
	Price uint
}

func main() {

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	ctx := context.Background()

	db.AutoMigrate(&Product{}) // Automatically creates indexes from struct tags
	
	// Only need explicit creation for conditional indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_price_created_d42 ON products(price, created_at) WHERE code = 'D42'")

	err = gorm.G[Product](db).Create(ctx, &Product{ID: uuid.New(), Code: "D42", Price: 100})
	if err != nil {
		log.Fatalf("failed to create product: %v", err)
	}

	err = gorm.G[Product](db).Create(ctx, &Product{ID: uuid.New(), Code: "D43", Price: 200})
	if err != nil {
		log.Fatalf("failed to create product: %v", err)
	}

	products, err := gorm.G[Product](db).Find(ctx)
	if err != nil {
		log.Fatalf("failed to find products: %v", err)
	}

	for _, product := range products {
		log.Printf("Product found: %+v\n", product)
	}

	product, err := gorm.G[Product](db).Where("id = ?", products[len(products)-1].ID).First(ctx) // find product with integer primary key
	if err != nil {
		log.Fatalf("failed to find product: %v", err)
	}

	log.Printf("Product found: %+v", product)
}
