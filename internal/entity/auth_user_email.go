package entity

import (
	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// OtherEmailHolders returns the number of other accounts that are not deleted and have the same email address,
// compared case-insensitively, and whether one of them has it verified.
func (m *User) OtherEmailHolders() (count int, verified bool, err error) {
	if m == nil {
		return 0, false, nil
	}

	email := clean.Email(m.UserEmail)

	if email == "" {
		return 0, false, nil
	}

	stmt := Db().Model(&User{}).Where("LOWER(user_email) = ? AND user_uid <> ?", email, m.UserUID)

	var n int64

	if err = stmt.Count(&n).Error; err != nil || n == 0 {
		return 0, false, err
	}

	count = int(n)

	var found []User

	if err = stmt.Select("verified_at").Where("verified_at IS NOT NULL").Find(&found).Error; err != nil {
		return count, false, err
	}

	for i := range found {
		if found[i].EmailVerified() {
			return count, true, nil
		}
	}

	return count, false, nil
}

// ReportSharedEmail adds a system log entry if other accounts hold the email address of the account,
// as a warning if one of them has it verified, and returns their number and whether it is verified there.
func (m *User) ReportSharedEmail() (count int, verified bool) {
	if count, verified, _ = m.OtherEmailHolders(); count == 0 {
		return count, verified
	}

	holders := english.Plural(count, "other account", "other accounts")

	if verified {
		event.SystemWarn([]string{"users", "%s", "email is also verified for %s"}, clean.LogQuote(m.UserName), holders)
	} else {
		event.SystemInfo([]string{"users", "%s", "email is also assigned to %s"}, clean.LogQuote(m.UserName), holders)
	}

	return count, verified
}
