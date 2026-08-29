package search

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/i18n"
)

// Errors a search may return, translated where the message reaches a user.
var (
	ErrForbidden    = i18n.Error(i18n.ErrForbidden)
	ErrBadRequest   = i18n.Error(i18n.ErrBadRequest)
	ErrNotFound     = i18n.Error(i18n.ErrNotFound)
	ErrBadSortOrder = fmt.Errorf("invalid sort order")
	ErrBadFilter    = fmt.Errorf("invalid search filter")
	ErrInvalidId    = fmt.Errorf("invalid ID specified")
)
