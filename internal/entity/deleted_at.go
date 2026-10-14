package entity

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// DeletedAt is the soft delete timestamp of a model, null while the row is not deleted.
// It is serialized like a *time.Time: null or omitted in JSON, a plain timestamp in YAML.
type DeletedAt sql.NullTime

// DeletedAtTime returns a DeletedAt set to the specified time.
func DeletedAtTime(t time.Time) DeletedAt {
	return DeletedAt{Time: t, Valid: true}
}

// IsZero reports whether the timestamp is not set, so that omitzero and omitempty skip it.
func (n DeletedAt) IsZero() bool {
	return !n.Valid
}

// Ptr returns the timestamp as a pointer, or nil if it is not set.
func (n DeletedAt) Ptr() *time.Time {
	if !n.Valid {
		return nil
	}

	t := n.Time

	return &t
}

// Scan implements the sql.Scanner interface.
func (n *DeletedAt) Scan(value any) error {
	return (*sql.NullTime)(n).Scan(value)
}

// Value implements the driver.Valuer interface.
func (n DeletedAt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return n.Time, nil
}

// MarshalJSON returns the timestamp as JSON, or null if it is not set.
func (n DeletedAt) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Time)
}

// UnmarshalJSON parses a JSON timestamp or null.
func (n *DeletedAt) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*n = DeletedAt{}
		return nil
	}

	var t time.Time

	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	*n = DeletedAtTime(t)

	return nil
}

// MarshalYAML returns the timestamp as a YAML scalar, or null if it is not set.
func (n DeletedAt) MarshalYAML() (any, error) {
	if !n.Valid {
		return nil, nil
	}

	return n.Time, nil
}

// UnmarshalYAML parses a YAML timestamp or null.
func (n *DeletedAt) UnmarshalYAML(unmarshal func(any) error) error {
	var t *time.Time

	if err := unmarshal(&t); err != nil {
		return err
	}

	if t == nil {
		*n = DeletedAt{}
	} else {
		*n = DeletedAtTime(*t)
	}

	return nil
}

// QueryClauses excludes soft-deleted rows from queries, as gorm.DeletedAt does.
func (n DeletedAt) QueryClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).QueryClauses(f)
}

// UpdateClauses excludes soft-deleted rows from updates, as gorm.DeletedAt does.
func (n DeletedAt) UpdateClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).UpdateClauses(f)
}

// DeleteClauses turns deletes into soft deletes, as gorm.DeletedAt does.
func (n DeletedAt) DeleteClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).DeleteClauses(f)
}
