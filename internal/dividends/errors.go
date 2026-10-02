package dividends

import "errors"

var ErrDividendNotFound = errors.New("dividend not found")
var ErrDividendAlreadyExists = errors.New("dividend already exists")
var ErrInvalidDividendYear = errors.New("invalid year")
var ErrInvalidDividendAmount = errors.New("amount must be greater than zero")
