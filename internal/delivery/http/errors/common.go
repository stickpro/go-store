package errors

import (
	"github.com/stickpro/go-store/internal/tools/apierror"
)

var ErrNoMatchesFound = apierror.Error{
	Message: "no matches found",
}
