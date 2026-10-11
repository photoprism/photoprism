package migrate

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// sqliteNumber is a value that a legacy binary column stored as a number, e.g. 1 for "0001".
type sqliteNumber struct {
	Key    string  `gorm:"column:number_key"`
	Column string  `gorm:"column:number_column"`
	Root   string  `gorm:"column:number_root"`
	Value  float64 `gorm:"column:number_value"`
	Int    *int64  `gorm:"column:number_int"`
	Text   string  `gorm:"column:number_text"`
}

// sqliteNumeric matches the decimal literals that SQLite stores as numbers in a column with numeric affinity.
var sqliteNumeric = regexp.MustCompile(`^[\s\v]*[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?[\s\v]*$`)

const (
	// maxExactFloat is the integer below which every integer has an exact float64 value.
	maxExactFloat = 1 << 53
	// maxInt64Float is the smallest float64 above the int64 range.
	maxInt64Float = 1 << 63
)

// sqliteRepairs restore the strings of path and name columns that were stored as numbers, using the
// file and folder names they were derived from; other numbers keep their converted text.
var sqliteRepairs = map[string]func(tx *gorm.DB, old string) error{
	"photos":  repairSQLitePhotos,
	"folders": repairSQLiteFolders,
	"albums":  repairSQLiteAlbums,
}

// sqliteNumbers returns the values of the columns that the old table stored as numbers.
func sqliteNumbers(tx *gorm.DB, old, key, root, where string, columns ...string) (result []sqliteNumber, err error) {
	for _, col := range columns {
		var rows []sqliteNumber

		query := fmt.Sprintf("SELECT CAST(%[1]s AS TEXT) AS number_key, ? AS number_column, %[2]s AS number_root, "+
			"%[3]s AS number_value, CASE WHEN typeof(%[3]s) = 'integer' THEN %[3]s END AS number_int, "+
			"CAST(%[3]s AS TEXT) AS number_text FROM %[4]s WHERE typeof(%[3]s) IN ('integer', 'real')%[5]s",
			sqliteIdent(key), root, sqliteIdent(col), sqliteIdent(old), where)

		if err = tx.Raw(query, col).Scan(&rows).Error; err != nil {
			return nil, err
		}

		result = append(result, rows...)
	}

	return result, nil
}

// Equals reports whether SQLite would have stored the candidate string as this number.
func (n sqliteNumber) Equals(candidate string) bool {
	if !sqliteNumeric.MatchString(candidate) {
		return false
	}

	c := strings.Trim(candidate, " \t\n\v\f\r")

	// Integers are compared exactly, as float64 cannot tell large integers apart.
	if n.Int != nil {
		if i, err := strconv.ParseInt(c, 10, 64); err == nil {
			return i == *n.Int
		} else if f, err := strconv.ParseFloat(c, 64); err != nil || f != float64(*n.Int) {
			return false
		} else {
			return *n.Int > -maxExactFloat && *n.Int < maxExactFloat || f < maxInt64Float && int64(f) == *n.Int
		}
	}

	f, err := strconv.ParseFloat(c, 64)

	return err == nil && f == n.Value
}

// sqliteMatch returns the candidate that equals the number, or false if none or several distinct ones do.
func sqliteMatch(n sqliteNumber, candidates []string) (match string, ok bool) {
	for _, c := range candidates {
		if !n.Equals(c) || c == match {
			continue
		} else if ok {
			return "", false
		}

		match, ok = c, true
	}

	return match, ok
}

// sqliteRestore updates a converted value with the string it was stored from, and returns false on error
// or if the converted text already is that string.
func sqliteRestore(tx *gorm.DB, table, key string, n sqliteNumber, value string) bool {
	if value == n.Text {
		return false
	} else if err := tx.Exec(fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", sqliteIdent(table), sqliteIdent(n.Column), sqliteIdent(key)),
		value, n.Key).Error; err != nil {
		log.Warnf("migrate: %s while restoring %s.%s %s (convert schema)", err, table, n.Column, clean.Log(value))
		return false
	}

	log.Debugf("migrate: restored %s.%s %s (convert schema)", table, n.Column, clean.Log(value))

	return true
}

// logRestored logs how many of the values that were stored as numbers have been restored.
func logRestored(table string, restored, found int) {
	if found > 0 {
		log.Infof("migrate: restored %d of %d names stored as numbers in %s (convert schema)", restored, found, table)
	}
}

// repairSQLitePhotos restores photo paths and names from the names of their files, primary files first.
func repairSQLitePhotos(tx *gorm.DB, old string) error {
	numbers, err := sqliteNumbers(tx, old, "id", "''", "", "photo_path", "photo_name")

	if err != nil {
		return err
	}

	restored := 0
	defer func() { logRestored("photos", restored, len(numbers)) }()

	for _, n := range numbers {
		var names []string

		if err = tx.Raw("SELECT file_name FROM files WHERE photo_id = ? AND typeof(file_name) = 'text' "+
			"ORDER BY file_primary DESC, id", n.Key).Scan(&names).Error; err != nil {
			return err
		}

		for _, name := range names {
			var candidates []string

			// A numeric file name may itself have been converted from a number.
			if _, numErr := strconv.ParseFloat(name, 64); numErr == nil {
				continue
			} else if n.Column == "photo_path" {
				if candidates = []string{path.Dir(name)}; candidates[0] == "." {
					candidates[0] = ""
				}
			} else {
				candidates = []string{fs.BasePrefix(name, false), fs.BasePrefix(name, true)}
			}

			if value, ok := sqliteMatch(n, candidates); ok {
				if sqliteRestore(tx, "photos", "id", n, value) {
					restored++
				}

				break
			}
		}
	}

	return nil
}

// sqliteTopFolders returns the distinct top-level folder names of a root found in file and folder paths
// that are stored as text. Only top-level names can be numbers, as other paths contain a slash.
func sqliteTopFolders(tx *gorm.DB, folders, root string) (result []string, err error) {
	err = tx.Raw(fmt.Sprintf("SELECT substr(file_name, 1, instr(file_name, '/') - 1) FROM files "+
		"WHERE file_root = ? AND typeof(file_name) = 'text' AND instr(file_name, '/') > 1 UNION "+
		"SELECT CASE WHEN instr(path, '/') > 0 THEN substr(path, 1, instr(path, '/') - 1) ELSE path END FROM %s "+
		"WHERE root = ? AND typeof(path) = 'text'", sqliteIdent(folders)), root, root).Scan(&result).Error

	return result, err
}

// repairSQLiteFolders restores top-level folder paths from the file and folder paths of the same root.
func repairSQLiteFolders(tx *gorm.DB, old string) error {
	numbers, err := sqliteNumbers(tx, old, "folder_uid", "root", "", "path")

	if err != nil {
		return err
	}

	tops := make(map[string][]string)
	restored := 0
	defer func() { logRestored("folders", restored, len(numbers)) }()

	for _, n := range numbers {
		if _, found := tops[n.Root]; !found {
			if tops[n.Root], err = sqliteTopFolders(tx, old, n.Root); err != nil {
				return err
			}
		}

		if value, ok := sqliteMatch(n, tops[n.Root]); ok && sqliteRestore(tx, "folders", "folder_uid", n, value) {
			restored++
		}
	}

	return nil
}

// repairSQLiteAlbums restores the paths of folder albums from the file and folder paths of the originals.
func repairSQLiteAlbums(tx *gorm.DB, old string) error {
	numbers, err := sqliteNumbers(tx, old, "album_uid", "''", " AND album_type = 'folder'", "album_path")

	if err != nil || len(numbers) == 0 {
		return err
	}

	tops, err := sqliteTopFolders(tx, "folders", "/")

	if err != nil {
		return err
	}

	restored := 0

	for _, n := range numbers {
		if value, ok := sqliteMatch(n, tops); ok && sqliteRestore(tx, "albums", "album_uid", n, value) {
			restored++
		}
	}

	logRestored("albums", restored, len(numbers))

	return nil
}
