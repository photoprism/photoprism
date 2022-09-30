package search

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/i18n"
)

var (
	ErrForbidden    = i18n.Error(i18n.ErrForbidden)
	ErrBadRequest   = i18n.Error(i18n.ErrBadRequest)
	ErrBadSortOrder = fmt.Errorf("invalid sort order")
	ErrBadFilter    = fmt.Errorf("invalid search filter")
	ErrInvalidId    = fmt.Errorf("invalid ID specified")
)
