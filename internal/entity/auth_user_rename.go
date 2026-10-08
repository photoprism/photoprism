package entity

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UserRenameHandler updates the copies of a username that a table keeps. It runs inside the
// rename transaction and must use tx, not the global connection.
type UserRenameHandler func(tx *gorm.DB, m *User, oldName, newName string) error

// UserRenameHandlers maps a table name to its handler; edition packages add theirs in init().
var UserRenameHandlers = map[string]UserRenameHandler{
	Session{}.TableName(): renameSessionUser,
	Client{}.TableName():  renameClientUser,
}

// CheckRename returns newName in its stored form if the account can be renamed to it, or an empty
// string if the name does not change. It applies the rules for a new account, limits the name to the
// length a login accepts, and refuses a name that another account, including a deleted one, holds.
func (m *User) CheckRename(newName string) (string, error) {
	if m.IsSystemOrInvalid() {
		return "", fmt.Errorf("system users cannot be modified")
	}

	name, err := authn.Username(newName)

	switch {
	case name != "" && name == clean.Username(m.UserName):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("username is %s", err.Error())
	case len(name) > txt.ClipUsername:
		return "", fmt.Errorf("username is %s", authn.ErrTooLong.Error())
	case len(name) < UsernameLength:
		return "", fmt.Errorf("username must have at least %d characters", UsernameLength)
	case clean.Handle(name) == "":
		return "", fmt.Errorf("username %s is not supported", clean.LogQuote(name))
	}

	if err = m.checkUsernameHolder(name); err != nil {
		return "", err
	}

	return name, nil
}

// checkUsernameHolder returns an error if another account, including a deleted one, holds the name.
func (m *User) checkUsernameHolder(name string) error {
	var holder User

	if err := UnscopedDb().Where("user_name = ? AND id <> ?", name, m.ID).First(&holder).Error; err == nil {
		return fmt.Errorf("user %s already exists", clean.LogQuote(name))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// deletedUsernameHolder reports whether a deleted account other than this one holds the name.
func (m *User) deletedUsernameHolder(name string) bool {
	var holder User

	return UnscopedDb().Where("user_name = ? AND id <> ? AND deleted_at IS NOT NULL", name, m.ID).First(&holder).Error == nil
}

// SetNewUsername checks newName with CheckRename and sets it, keeping the folders that derive from
// the current name, and sets an empty display name. If the name does not change, it sets the
// normalized form of a stored name instead. SaveRename writes the change.
func (m *User) SetNewUsername(newName string) error {
	name, err := m.CheckRename(newName)

	if err != nil {
		return err
	} else if name == "" {
		m.SetNormalizedUsername()
		return nil
	}

	m.setRenamedUsername(name)

	if m.DisplayName == "" {
		m.DisplayName = m.FullName()
	}

	return nil
}

// SetNormalizedUsername sets the stored username in the form Validate normalizes it to, keeping the
// folders that derive from it, and reports whether it changed. It keeps a name that another account,
// including a deleted one, holds in its normalized form.
func (m *User) SetNormalizedUsername() bool {
	name, err := m.normalizedUsername()

	if err != nil || name == m.UserName || m.IsSystemOrInvalid() || m.checkUsernameHolder(name) != nil {
		return false
	}

	m.setRenamedUsername(name)

	return true
}

// setRenamedUsername pins the paths that resolve from the current name if the new name resolves them
// differently, then sets the name.
func (m *User) setRenamedUsername(name string) {
	renamed := User{UserUID: m.UserUID, UserName: name}

	if m.DefaultBasePath() != renamed.DefaultBasePath() {
		m.GetBasePath()

		if m.BasePath == "" && m.UploadPath == "~" && m.UserName != "" {
			m.UploadPath = m.DefaultBasePath()
		}
	}

	m.UserName = name
}

// Rename changes the username of the account and every stored copy of it in one transaction.
func (m *User) Rename(newName string) error {
	oldName := m.UserName

	if err := m.SetNewUsername(newName); err != nil {
		return err
	} else if m.UserName == oldName {
		return nil
	}

	return m.SaveRename(oldName)
}

// SaveRename saves the account and replaces oldName in the tables of UserRenameHandlers in one
// transaction. It saves the account like Save if the name has not changed.
func (m *User) SaveRename(oldName string) error {
	if m.UserName == oldName {
		return m.Save()
	} else if m.IsSystemOrInvalid() {
		return fmt.Errorf("system users cannot be modified")
	}

	m.GenerateTokens(false)

	err := Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Save(m).Error; err != nil {
			return err
		}

		for _, table := range slices.Sorted(maps.Keys(UserRenameHandlers)) {
			if err := UserRenameHandlers[table](tx, m, oldName, m.UserName); err != nil {
				return fmt.Errorf("failed to update %s (%w)", table, err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	m.SaveRelated()
	FlushUserSessionCache(m.UserUID)

	return nil
}

// renameSessionUser updates the username stored with the sessions of the account.
func renameSessionUser(tx *gorm.DB, m *User, _, newName string) error {
	res := tx.Model(&Session{}).Where("user_uid = ?", m.UserUID).UpdateColumn("user_name", newName)

	if res.Error != nil {
		return res.Error
	}

	log.Debugf("users: changed username of %s", english.Plural(int(res.RowsAffected), "session", "sessions"))

	return nil
}

// renameClientUser updates the username stored with the clients of the account, including deleted ones.
func renameClientUser(tx *gorm.DB, m *User, _, newName string) error {
	res := tx.Unscoped().Model(&Client{}).Where("user_uid = ?", m.UserUID).UpdateColumn("user_name", newName)

	if res.Error != nil {
		return res.Error
	}

	log.Debugf("users: changed username of %s", english.Plural(int(res.RowsAffected), "client", "clients"))

	return nil
}
