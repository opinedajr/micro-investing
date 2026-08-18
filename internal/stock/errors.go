package stock

import "errors"

var (
	ErrStockNotFound     = errors.New("stock not found")
	ErrInvalidTicker     = errors.New("invalid ticker format")
	ErrInvalidName       = errors.New("invalid name length")
	ErrInvalidSector     = errors.New("invalid sector length")
	ErrInvalidRank       = errors.New("invalid rank value")
	ErrFailedToList      = errors.New("failed to list stocks")
	ErrFailedToCreate     = errors.New("failed to create stock")
	ErrFailedToFind       = errors.New("failed to find stock")
	ErrFailedToSeed      = errors.New("failed to seed stocks")
)
