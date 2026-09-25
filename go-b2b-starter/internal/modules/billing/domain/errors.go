package domain

import "errors"

var ErrCheckoutOwnership = errors.New("checkout does not belong to this organization")
var ErrPaymentPending = errors.New("payment is not completed")
