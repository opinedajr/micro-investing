package position

import "errors"

var ErrPositionNotFound = errors.New("position not found")
var ErrPositionAlreadyExists = errors.New("position already exists")
var ErrInvalidPositionQuantity = errors.New("quantity must be greater than zero")
var ErrInvalidPositionAveragePrice = errors.New("average_price must be greater than zero")
var ErrStockNotFound = errors.New("stock not found")
