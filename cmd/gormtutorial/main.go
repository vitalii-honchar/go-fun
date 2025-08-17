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
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Code  string
	Price uint
}

func main() {

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	ctx := context.Background()

	db.AutoMigrate(&Product{})

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
