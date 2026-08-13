package patrimony

import "errors"

var ErrPatrimonyAlreadyExists = errors.New("patrimony already exists")
var ErrPatrimonyNotFound = errors.New("patrimony not found")
var ErrInvalidAssetType = errors.New("invalid asset type")
var ErrInvalidPatrimonyYear = errors.New("invalid year")
var ErrInvalidPatrimonyMonth = errors.New("invalid month")
var ErrInvalidPatrimonyAmount = errors.New("amount cannot be negative")
var ErrAssetNotFound = errors.New("asset not found")
var ErrInvalidAssetDate = errors.New("asset date cannot be in the future")
var ErrInvalidAssetDescription = errors.New("asset description must be between 3 and 100 characters")
var ErrInvalidAssetAmount = errors.New("asset amount must be greater than zero")
var ErrInvalidDateRange = errors.New("start_date must be on or before end_date")
var ErrInvalidFilterDate = errors.New("invalid date format for filter parameter")
