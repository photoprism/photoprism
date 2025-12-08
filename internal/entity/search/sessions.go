package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Sessions finds user sessions.
func Sessions(frm form.SearchSessions) (result entity.Sessions, err error) {
	result = entity.Sessions{}
	stmt := Db()

	userUid := strings.TrimSpace(frm.UID)
	search := strings.TrimSpace(frm.Query)

	order := frm.Order
	limit := frm.Count
	offset := frm.Offset

	// Limit maximum number of results.
	if limit > MaxResults {
		limit = MaxResults
	}

	// Set default sort order or use normalized order value.
	if order == "" {
		order = sortby.LastActive
	} else {
		order = clean.TypeLowerUnderscore(order)
	}

	// Filter by user UID?
	if userUid != "" {
		stmt = stmt.Where("user_uid = ?", userUid)
	}

	// Filter by username and/or auth provider name?
	if search != "" && search != "all" {
		stmt = stmt.Where("user_name LIKE ? OR client_name LIKE ?", search+"%", search+"%")
	}

	// Filter by authentication providers?
	if frm.Provider != "" {
		stmt = stmt.Where("auth_provider IN (?)", frm.AuthProviders())
	}

	// Filter by authentication methods?
	if frm.Method != "" {
		stmt = stmt.Where("auth_method IN (?)", frm.AuthMethods())
	}

	// Sort results?
	switch order {
	case sortby.LastActive:
		stmt = stmt.Order(OrderExpr("last_active DESC, user_name, client_name, id", frm.Reverse))
	case sortby.SessExpires:
		stmt = stmt.Order(OrderExpr("sess_expires DESC, user_name, client_name, id", frm.Reverse))
	case sortby.ClientName:
		stmt = stmt.
			Where("client_name <> '' AND client_name IS NOT NULL").
			Order(OrderExpr("client_name, created_at, id", frm.Reverse))
	case sortby.Login, sortby.LoginAt:
		stmt = stmt.Order(OrderExpr("login_at DESC, user_name, client_name, id", frm.Reverse))
	case sortby.Created, sortby.CreatedAt:
		stmt = stmt.Order(OrderExpr("created_at ASC, user_name, client_name, id", frm.Reverse))
	case sortby.Updated, sortby.UpdatedAt:
		stmt = stmt.Order(OrderExpr("updated_at DESC, user_name, client_name, id", frm.Reverse))
	default:
		return result, fmt.Errorf("invalid sort order %s", order)
	}

	// Apply limit and offset.
	if limit > 0 {
		stmt = stmt.Limit(limit)

		if offset > 0 {
			stmt = stmt.Offset(offset)
		}
	}

	// Perform query.
	err = stmt.Find(&result).Error

	return result, err
}
