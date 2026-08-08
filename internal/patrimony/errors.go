package patrimony

import "errors"

var ErrPatrimonyAlreadyExists = errors.New("patrimony already exists")
var ErrPatrimonyNotFound = errors.New("patrimony not found")
var ErrInvalidAssetType = errors.New("invalid asset type")
var ErrInvalidPatrimonyYear = errors.New("invalid year")
var ErrInvalidPatrimonyMonth = errors.New("invalid month")
var ErrInvalidPatrimonyAmount = errors.New("amount cannot be negative")
