package dto

import "github.com/google/uuid"

// Owner identifies the subject of a request — either an authenticated user or a guest session.
type Owner struct {
	UserID    *uuid.UUID
	SessionID *uuid.UUID
}
