package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Users finds user accounts based on the specified search parameters.
func Users(frm form.SearchUsers) (result entity.Users, err error) {
	result = entity.Users{}

	// Parse query string and filter.
	if err = frm.ParseQueryString(); err != nil {
		log.Debugf("users: %s", err)
		return result, ErrBadRequest
	}

	// Set search filters based on search terms.
	if terms := txt.SearchTerms(frm.Query); frm.Query != "" && len(terms) > 0 {
		switch {
		case terms["all"]:
			frm.Query = strings.ReplaceAll(frm.Query, "all", "")
			frm.All = true
		case terms["deleted"]:
			frm.Query = strings.ReplaceAll(frm.Query, "deleted", "")
			frm.Deleted = true
		}
	}

	stmt := UnscopedDb()

	search := strings.TrimSpace(frm.Query)
	sortOrder := frm.Order
	limit := frm.Count
	offset := frm.Offset

	if frm.All {
		// Don't filter.
	} else if id := txt.Int(search); id != 0 {
		stmt = stmt.Where("id = ?", id)
	} else if rnd.IsUID(search, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		stmt = stmt.Where("user_name LIKE ? OR user_email LIKE ? OR display_name LIKE ?", search+"%", search+"%", search+"%")
	} else {
		stmt = stmt.Where("id > 0")
	}

	// Find deleted user accounts?
	if frm.Deleted {
		stmt = stmt.Where("deleted_at IS NOT NULL")
	} else if !frm.All {
		stmt = stmt.Where("deleted_at IS NULL")
	}

	switch sortOrder {
	case sortby.Name:
		sortOrder = "user_name, id"
	case sortby.DisplayName:
		sortOrder = "display_name, id"
	case sortby.Login, sortby.LoginAt:
		sortOrder = "login_at DESC, id"
	case sortby.Created, sortby.CreatedAt:
		sortOrder = "created_at ASC, id"
	case sortby.Updated, sortby.UpdatedAt:
		sortOrder = "updated_at DESC, id"
	case sortby.Deleted, sortby.DeletedAt:
		sortOrder = "deleted_at DESC, created_at DESC, id"
	case sortby.Email:
		sortOrder = "user_email, id"
	default:
		sortOrder = "user_name, id"
	}

	if limit > 0 {
		stmt = stmt.Limit(limit)

		if offset > 0 {
			stmt = stmt.Offset(offset)
		}
	}

	err = stmt.Order(sortOrder).Find(&result).Error

	return result, err
}
