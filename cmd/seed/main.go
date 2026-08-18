package main

import (
	"context"
	"flag"
	"log"

	"github.com/opinedajr/micro-investing/internal/di"
)

func main() {
	force := flag.Bool("force", false, "Update existing stocks instead of skipping them")
	flag.Parse()

	container := di.NewContainer()

	inserted, updated, skipped, err := container.StockService().SeedStocks(context.Background(), *force)
	if err != nil {
		log.Fatalf("Failed to seed stocks: %v", err)
	}

	log.Printf("Stock seeding completed: inserted=%d updated=%d skipped=%d", inserted, updated, skipped)
}
