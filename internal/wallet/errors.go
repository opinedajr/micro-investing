package wallet

import "errors"

var ErrWalletNameAlreadyExists = errors.New("wallet name already exists")
var ErrWalletNotFound = errors.New("wallet not found")
