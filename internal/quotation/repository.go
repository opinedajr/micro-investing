package quotation

import "context"

type Repository interface {
	UpsertCurrentPrices(ctx context.Context, prices []CurrentPrice) error
}
