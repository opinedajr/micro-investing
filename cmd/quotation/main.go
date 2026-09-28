package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/quotation"
	"github.com/opinedajr/micro-investing/internal/shared"
	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func main() {
	tickersFlag := flag.String("tickers", "", "Comma-separated tickers to sync (default: all registered stocks)")
	flag.Parse()

	defer exitOnPanic()

	ctx := context.Background()
	container := di.NewContainer()
	logger := container.Logger()

	syncOutput, err := container.QuotationService().SyncCurrentPrices(ctx, quotation.SyncInput{Tickers: parseTickers(*tickersFlag)})
	if err != nil {
		logger.Error(ctx, "current prices sync failed", "error", err)
		os.Exit(1)
	}

	warnProblematicTickers(ctx, logger, syncOutput)

	wallets, err := container.WalletService().List(ctx, shared.DefaultUserID)
	if err != nil {
		logger.Error(ctx, "failed to list wallets for consolidation", "error", err)
		os.Exit(1)
	}

	consolidationFailures := consolidateWallets(ctx, container, logger, wallets)

	fmt.Printf("current prices sync summary: updated=%d skipped=%d failed=%d wallets=%d consolidation_failures=%d\n",
		syncOutput.Updated, len(syncOutput.Skipped), len(syncOutput.Failed), len(wallets), len(consolidationFailures))

	if len(consolidationFailures) > 0 {
		os.Exit(1)
	}

	fmt.Println("current prices sync completed successfully")
}

func exitOnPanic() {
	if recovered := recover(); recovered != nil {
		fmt.Fprintf(os.Stderr, "failed to sync current prices: %v\n", recovered)
		os.Exit(1)
	}
}

func parseTickers(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	tickers := make([]string, 0, len(parts))
	for _, part := range parts {
		ticker := strings.TrimSpace(part)
		if ticker == "" {
			continue
		}
		tickers = append(tickers, ticker)
	}
	return tickers
}

func warnProblematicTickers(ctx context.Context, logger logger.Logger, syncOutput *quotation.SyncOutput) {
	for _, ticker := range syncOutput.Skipped {
		logger.Warn(ctx, "ticker skipped by provider", "ticker", ticker)
	}
	for _, ticker := range syncOutput.Failed {
		logger.Warn(ctx, "ticker failed to sync", "ticker", ticker)
	}
}

func consolidateWallets(ctx context.Context, container *di.Container, logger logger.Logger, wallets []wallet.WalletOutput) []string {
	failures := make([]string, 0)
	for _, walletItem := range wallets {
		if err := container.PositionService().ConsolidateByWallet(ctx, walletItem.ID); err != nil {
			logger.Error(ctx, "failed to consolidate wallet positions", "wallet_id", walletItem.ID, "wallet_name", walletItem.Name, "error", err)
			failures = append(failures, walletItem.ID)
			continue
		}
		logger.Info(ctx, "wallet positions consolidated", "wallet_id", walletItem.ID, "wallet_name", walletItem.Name)
	}
	return failures
}
