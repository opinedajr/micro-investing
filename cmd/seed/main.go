package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/stock"
)

func main() {
	force := flag.Bool("force", false, "Overwrite existing stocks on conflict")
	flag.Parse()

	container := di.NewContainer()
	service := container.StockService()

	if err := service.Seed(context.Background(), stock.SeedInput{Force: *force}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to seed stocks: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("stocks seeded successfully")
}
