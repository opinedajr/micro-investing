package quotation

import "errors"

var ErrQuotesFetchFailed = errors.New("quotes fetch failed")
var ErrUnknownTicker = errors.New("unknown ticker")
