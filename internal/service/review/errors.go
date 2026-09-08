package review

import "errors"

// ErrNotPurchased is returned by CreateProductReview when the author has no
// paid, non-cancelled order containing the variant being reviewed.
var ErrNotPurchased = errors.New("review: variant was not purchased by this user")
