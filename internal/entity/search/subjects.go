package search

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// sessionGrantsPeople reports whether the session is granted perm on people, mirroring
// sessionGrantsPhotos: a client is limited by its own role as well as its user's, and one without a
// user is evaluated on that role alone - reading the user role there resolves to RoleNone and denies
// what the request was admitted with. A nil session is internal or CLI use and is not restricted.
func sessionGrantsPeople(sess *entity.Session, perm acl.Permission) bool {
	if sess == nil {
		return true
	}

	if sess.IsClient() {
		if !acl.Rules.Allow(acl.ResourcePeople, sess.GetClientRole(), perm) {
			return false
		} else if sess.NoUser() {
			return true
		}
	}

	return acl.Rules.Allow(acl.ResourcePeople, sess.GetUserRole(), perm)
}

// SubjectSessionSeesPrivate reports whether a session may see people marked private, so the search
// and the handlers that read a single subject answer the same question about the same session.
func SubjectSessionSeesPrivate(sess *entity.Session) bool {
	return sessionGrantsPeople(sess, acl.AccessPrivate)
}

// Subjects searches subjects and returns them without checking rights or permissions.
func Subjects(frm form.SearchSubjects) (results SubjectResults, err error) {
	return searchSubjects(frm, nil)
}

// UserSubjects searches subjects within what the given session is allowed to see.
func UserSubjects(frm form.SearchSubjects, sess *entity.Session) (results SubjectResults, err error) {
	return searchSubjects(frm, sess)
}

// searchSubjects searches subjects and returns them, applying the session's own limits when one is
// given: a role without AccessPrivate must not reach a private person through any filter.
func searchSubjects(frm form.SearchSubjects, sess *entity.Session) (results SubjectResults, err error) {
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	// Check session permissions and apply as needed.
	if !SubjectSessionSeesPrivate(sess) {
		// Cleared as well as forced, since "all" skips the visibility filters wholesale.
		frm.Private = "no"
		frm.All = false
	}

	// Check session permissions and apply as needed.
	if !SubjectSessionSeesPrivate(sess) {
		// Cleared as well as forced, since "all" skips the visibility filters wholesale.
		frm.Private = "no"
		frm.All = false
	}

	results = make(SubjectResults, 0)
	subjTable := entity.Subject{}.TableName()

	// Base query.
	s := UnscopedDb().Table(subjTable).
		Select(fmt.Sprintf("%s.*", subjTable))

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Set sort order.
	switch frm.Order {
	case "name":
		s = s.Order(OrderExpr("subj_name ASC", frm.Reverse))
	case "count":
		s = s.Order(OrderExpr("file_count DESC", frm.Reverse))
	case "added":
		s = s.Order(OrderExpr(fmt.Sprintf("%s.created_at DESC", subjTable), frm.Reverse))
	case "relevance":
		if entity.DbDialect() == dsn.DialectPostgreSQL {
			s = s.Order(OrderExpr("subj_favorite DESC, photo_count DESC NULLS LAST", frm.Reverse))
		} else {
			s = s.Order(OrderExpr("subj_favorite DESC, photo_count DESC", frm.Reverse))
		}
	default:
		s = s.Order(OrderExpr("subj_favorite DESC, subj_name ASC", frm.Reverse))
	}

	// Applied before the uid shortcut below, so no branch can answer with a row the caller is not
	// allowed to see. The shortcut skips the content filters on purpose - a caller reloading one
	// person by uid wants that row whatever its file count - but never the visibility ones.
	s = s.Where(fmt.Sprintf("%s.deleted_at IS NULL", subjTable))

	if !frm.All {
		if txt.Yes(frm.Favorite) {
			s = s.Where("subj_favorite = 1")
		} else if txt.No(frm.Favorite) {
			s = s.Where("subj_favorite = 0")
		}

		if !txt.Yes(frm.Hidden) {
			s = s.Where("subj_hidden = 0")
		}

		if txt.Yes(frm.Private) {
			s = s.Where("subj_private = 1")
		} else if txt.No(frm.Private) {
			s = s.Where("subj_private = 0")
		}

		if txt.Yes(frm.Excluded) {
			s = s.Where("subj_excluded = 1")
		} else if txt.No(frm.Excluded) {
			s = s.Where("subj_excluded = 0")
		}
	}

	if frm.UID != "" {
		s = s.Where(fmt.Sprintf("%s.subj_uid IN (?)", subjTable), strings.Split(strings.ToLower(frm.UID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if frm.Query != "" {
		var wheres []string
		var values [][]any
		switch entity.DbDialect() {
		case dsn.DialectPostgreSQL:
			wheres, values = LikeAllNames(Cols{"lower(subj_name)", "lower(subj_alias)"}, strings.ToLower(frm.Query))
		default:
			wheres, values = LikeAllNames(Cols{"subj_name", "subj_alias"}, frm.Query)
		}
		for i, where := range wheres {
			s = s.Where("?", gorm.Expr(where, values[i]...))
		}
	}

	if frm.Files > 0 {
		s = s.Where("file_count >= ?", frm.Files)
	}

	if frm.Photos > 0 {
		s = s.Where("photo_count >= ?", frm.Photos)
	}

	if frm.Type != "" {
		s = s.Where("subj_type IN (?)", strings.Split(frm.Type, txt.Or))
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}

// SubjectUIDs finds subject UIDs matching the search string, and removes names from the remaining query.
func SubjectUIDs(s string) (result []string, names []string, remaining string) {
	if s == "" {
		return result, names, s
	}

	type Matches struct {
		SubjUID   string
		SubjName  string
		SubjAlias string
	}

	var matches []Matches
	whereString1 := ""
	whereString2 := ""
	valueString := ""
	switch entity.DbDialect() {
	case dsn.DialectPostgreSQL:
		whereString1 = "lower(subj_name)"
		whereString2 = "lower(subj_alias)"
		valueString = strings.ToLower(s)
	default:
		whereString1 = "subj_name"
		whereString2 = "subj_alias"
		valueString = s
	}
	wheres, values := LikeAllNames(Cols{whereString1, whereString2}, valueString)

	if len(wheres) == 0 {
		return result, names, s
	}

	remaining = s

	for i, where := range wheres {
		var subj []string

		stmt := Db().Model(&entity.Subject{})
		stmt = stmt.Where("?", gorm.Expr(where, values[i]...))

		if err := stmt.Scan(&matches).Error; err != nil {
			log.Errorf("search: %s while finding subjects", err)
		} else if len(matches) == 0 {
			continue
		}

		for _, m := range matches {
			subj = append(subj, m.SubjUID)
			names = append(names, m.SubjName)

			for _, r := range txt.Words(strings.ToLower(m.SubjName)) {
				if len(r) > 1 {
					remaining = strings.ReplaceAll(remaining, r, "")
				}
			}

			for _, r := range txt.Words(strings.ToLower(m.SubjAlias)) {
				if len(r) > 1 {
					remaining = strings.ReplaceAll(remaining, r, "")
				}
			}
		}

		result = append(result, strings.Join(subj, txt.Or))
	}

	return result, names, clean.SearchQuery(remaining)
}
