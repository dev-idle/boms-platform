// Command seed fills a development database with the demo catalog: categories, products,
// combos and discount codes. It is idempotent — rows already present are left untouched.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()
	pool, err := postgresrepo.NewPool(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()

	report, err := bootstrap.SeedCatalog(ctx, cfg, bootstrap.CatalogSeedDeps{
		Categories: postgresrepo.NewCategoryRepository(pool),
		Products:   postgresrepo.NewProductRepository(pool),
		Combos:     postgresrepo.NewComboRepository(pool),
		Discounts:  postgresrepo.NewDiscountCodeRepository(pool),
	}, time.Now().UTC())
	if err != nil {
		return err
	}

	fmt.Printf(
		"seeded: %d categories, %d products, %d combos, %d discount codes\n",
		report.Categories, report.Products, report.Combos, report.DiscountCodes,
	)
	return nil
}
