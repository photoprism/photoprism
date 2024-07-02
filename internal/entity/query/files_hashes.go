package query

import (
	"database/sql"

	"github.com/photoprism/photoprism/internal/entity"
)

type HashMap map[string]bool

// CountFileHashes counts distinct file hashes.
func CountFileHashes() (count int) {
	if err := UnscopedDb().
		Table(entity.File{}.TableName()).
		Where("file_missing = 0 AND deleted_at IS NULL").
		Select("COUNT(DISTINCT(file_hash))").Count(&count).Error; err != nil {
		log.Errorf("files: %s (count hashes)", err)
	}

	return count
}

// FetchHashMap populates a hash map from the database.
func FetchHashMap(rows *sql.Rows, result HashMap, hashLen int) (err error) {
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	for rows.Next() {
		var h string

		if err = rows.Scan(&h); err != nil {
			return err
		} else if len(h) > hashLen {
			result[h[:hashLen]] = true
		} else if h != "" {
			result[h] = true
		}
	}

	return nil
}

// FileHashMap returns a map of all known file hashes.
func FileHashMap() (result HashMap, err error) {
	count := CountFileHashes()

	result = make(HashMap, count)

	if rows, err := UnscopedDb().
		Table(entity.File{}.TableName()).
		Where("file_missing = 0 AND deleted_at IS NULL").
		Where("file_hash IS NOT NULL AND file_hash <> ''").
		Select("file_hash").Rows(); err != nil {
		return result, err
	} else if err := FetchHashMap(rows, result, 40); err != nil {
		return result, err
	}

	return result, err
}

// ThumbHashMap returns a map of all known thumb file hashes.
func ThumbHashMap() (result HashMap, err error) {
	tables := []string{
		entity.Album{}.TableName(),
		entity.Label{}.TableName(),
		entity.Marker{}.TableName(),
		entity.Subject{}.TableName(),
		entity.User{}.TableName(),
	}

	result = make(HashMap)

	for i := range tables {
		if rows, err := UnscopedDb().
			Table(tables[i]).
			Where("thumb IS NOT NULL AND thumb <> ''").
			Select("thumb").Rows(); err != nil {
			return result, err
		} else if err := FetchHashMap(rows, result, 40); err != nil {
			return result, err
		}
	}

	return result, nil
}
