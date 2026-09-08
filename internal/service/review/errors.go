package review

import "errors"

// ErrNotPurchased — the user tried to review a variant they have not bought.
// A purchase counts only once its order reached a paid-or-later status
// (paid / processing / shipped / delivered).
var ErrNotPurchased = errors.New("review is only allowed for a purchased product")
