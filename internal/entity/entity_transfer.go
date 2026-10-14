package entity

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TransferBatchSize is the default number of rows that Tables.Transfer copies per statement.
const TransferBatchSize = 1000

// transferMaxVars limits the values bound per insert statement, below the limits of SQLite, MariaDB and PostgreSQL.
const transferMaxVars = 30000

// TransferOptions configures a transfer of rows from one database to another.
type TransferOptions struct {
	// BatchSize is the number of rows copied per statement.
	BatchSize int
	// Progress is called after each table with the number of rows copied, if it is not nil.
	Progress func(table string, rows int64)
}

// DataTables returns the names of the tables that hold data, i.e. all except the schemaTables, in alphabetical order.
func (list Tables) DataTables() []string {
	names := make([]string, 0, len(list))

	for name := range list {
		if !schemaTables[name] {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	return names
}

// Rows returns the number of rows in each existing data table of db that is not empty.
func (list Tables) Rows(db *gorm.DB) (map[string]int64, error) {
	result := make(map[string]int64)

	for _, name := range list.DataTables() {
		if !db.Migrator().HasTable(name) {
			continue
		}

		var count int64

		if err := db.Table(name).Count(&count).Error; err != nil {
			return result, fmt.Errorf("%s (count rows of %s)", err, name)
		} else if count > 0 {
			result[name] = count
		}
	}

	return result, nil
}

// Transfer replaces the rows of the data tables in dst with those in src and returns the number of rows copied
// per table. Both schemas must be migrated by the same version. Rows are copied as stored, without hooks or defaults,
// from one snapshot of src. If the transfer does not complete, the data tables in dst are cleared again.
func (list Tables) Transfer(src, dst *gorm.DB, opt TransferOptions) (result map[string]int64, err error) {
	if src == nil || dst == nil {
		return nil, errors.New("database not connected")
	}

	if opt.BatchSize < 1 {
		opt.BatchSize = TransferBatchSize
	}

	names := list.DataTables()
	result = make(map[string]int64, len(names))
	autoInc := autoIncrementTables(dst)

	for _, name := range names {
		if err = truncateTable(dst, name, autoInc[name]); err != nil {
			return result, fmt.Errorf("%s (clear %s)", err, name)
		}
	}

	// The target is cleared on every way out except success, including a panic.
	done := false

	defer func() {
		if done {
			return
		}

		for _, name := range names {
			if clearErr := truncateTable(dst, name, autoInc[name]); clearErr != nil {
				log.Warnf("transfer: %s (clear %s)", clearErr, name)
			}
		}
	}()

	// The source is read in one transaction, so all tables are copied as they were at the same time.
	tx := src.Begin(transferTxOptions(src))

	if tx.Error != nil {
		return result, fmt.Errorf("%s (begin transaction)", tx.Error)
	}

	defer tx.Rollback()

	// Row counts and the latest update times in the snapshot, to detect writes to the source after it.
	snapshot := make(map[string]transferState, len(names))

	for _, name := range names {
		var rows, skipped int64

		if rows, skipped, err = transferTable(tx, dst, name, list[name], opt.BatchSize); err != nil {
			return result, fmt.Errorf("%s (copy %s)", err, name)
		} else if snapshot[name], err = transferStateOf(tx, name, list[name]); err != nil {
			return result, err
		} else if snapshot[name].Rows != rows+skipped {
			return result, fmt.Errorf("%d of %d rows copied from %s", rows+skipped, snapshot[name].Rows, name)
		} else if skipped > 0 {
			log.Warnf("transfer: skipped %s of %s whose keys the target treats as already present, without merging them", english.Plural(int(skipped), "row", "rows"), name)
		}

		result[name] = rows

		if opt.Progress != nil {
			opt.Progress(name, rows)
		}
	}

	if err = SyncSequences(dst); err != nil {
		return result, fmt.Errorf("%s (sync sequences)", err)
	}

	var changed []string

	for _, name := range names {
		var count int64

		// The copied rows must all be in the target.
		if err = dst.Table(name).Count(&count).Error; err != nil {
			return result, fmt.Errorf("%s (count rows of %s)", err, name)
		} else if count != result[name] {
			return result, fmt.Errorf("%d of %d rows found in %s", count, result[name], name)
		}

		// Rows written to the source after its snapshot are missing in the target, except log entries.
		if TransferLogTables[name] {
			continue
		} else if state, stateErr := transferStateOf(src, name, list[name]); stateErr != nil {
			return result, stateErr
		} else if state != snapshot[name] {
			changed = append(changed, name)
		}
	}

	if len(changed) > 0 {
		return result, fmt.Errorf("the index changed during the transfer (%s)", strings.Join(changed, ", "))
	}

	done = true

	return result, nil
}

// TransferLogTables contains the log tables that Tables.Transfer does not check for writes after its snapshot,
// as the transfer itself may log to them.
var TransferLogTables = map[string]bool{
	Error{}.TableName(): true,
}

// TransferSkipDuplicates contains the log tables whose rows Tables.Transfer skips with a warning if their key is
// already in the target, e.g. because the target compares keys that differ only by case or accents as equal.
var TransferSkipDuplicates = map[string]bool{}

// isDuplicateKeyError reports whether a write failed because a row with the same unique key exists.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError

	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	var sqliteErr sqlite3.Error

	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}

	return false
}

// transferState holds the number of rows in a table and the latest update time, if it has an updated_at column.
type transferState struct {
	Rows    int64
	Updated string
}

// transferStateOf returns the number of rows and the latest update time of a table.
func transferStateOf(db *gorm.DB, name string, model any) (state transferState, err error) {
	if err = db.Table(name).Count(&state.Rows).Error; err != nil {
		return state, fmt.Errorf("%s (count rows of %s)", err, name)
	}

	stmt := &gorm.Statement{DB: db}

	if err = stmt.Parse(model); err != nil {
		return state, err
	} else if _, ok := stmt.Schema.FieldsByDBName["updated_at"]; !ok {
		return state, nil
	}

	var updated sql.NullString

	if err = db.Table(name).Select("MAX(updated_at)").Row().Scan(&updated); err != nil {
		return state, fmt.Errorf("%s (latest update of %s)", err, name)
	}

	state.Updated = updated.String

	return state, nil
}

// transferTxOptions returns the options of the transaction that reads the source of a transfer. SQLite holds
// its snapshot until the transaction ends and accepts no options.
func transferTxOptions(db *gorm.DB) *sql.TxOptions {
	if dsn.DialectFromDriver(db.Name()) == dsn.DialectSQLite {
		return nil
	}

	return &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}
}

// Unknown returns the names of the tables in db that are neither registered nor deprecated, e.g. those of
// another edition or application.
func (list Tables) Unknown(db *gorm.DB) ([]string, error) {
	tables, err := db.Migrator().GetTables()

	if err != nil {
		return nil, err
	}

	var result []string

	for _, name := range tables {
		if strings.HasPrefix(name, "sqlite_") {
			// Internal SQLite table, e.g. sqlite_sequence.
			continue
		} else if _, ok := list[name]; !ok && !slices.Contains(DeprecatedTables, name) {
			result = append(result, name)
		}
	}

	slices.Sort(result)

	return result, nil
}

// transferTable copies the rows of a table from src to dst in batches, ordered by primary key, and returns the number
// of rows copied and of rows skipped, as rows of TransferSkipDuplicates tables whose key is already in dst are skipped.
func transferTable(src, dst *gorm.DB, name string, model any, batchSize int) (total, skipped int64, err error) {
	stmt := &gorm.Statement{DB: src}

	if err = stmt.Parse(model); err != nil {
		return 0, 0, err
	}

	s := stmt.Schema

	if len(s.PrimaryFields) == 0 {
		return 0, 0, errors.New("no primary key")
	}

	fields := make([]*schema.Field, 0, len(s.DBNames))
	columns := make([]clause.Column, 0, len(s.DBNames))

	for _, dbName := range s.DBNames {
		fields = append(fields, s.FieldsByDBName[dbName])
		columns = append(columns, clause.Column{Name: dbName})
	}

	if maxRows := transferMaxVars / len(columns); batchSize > maxRows {
		batchSize = max(maxRows, 1)
	}

	order := clause.OrderBy{Columns: make([]clause.OrderByColumn, len(s.PrimaryFields))}

	for i, f := range s.PrimaryFields {
		order.Columns[i] = clause.OrderByColumn{Column: clause.Column{Name: f.DBName}}
	}

	read := src.Session(&gorm.Session{NewDB: true, SkipHooks: true})
	write := dst.Session(&gorm.Session{NewDB: true, SkipHooks: true})
	sliceType := reflect.SliceOf(reflect.PointerTo(s.ModelType))

	var last []any

	for {
		batch := reflect.New(sliceType)
		query := read.Table(name).Unscoped().Order(order).Limit(batchSize)

		if last != nil {
			cond, args := transferKeyset(s.PrimaryFields, last)
			query = query.Where(cond, args...)
		}

		if err = query.Find(batch.Interface()).Error; err != nil {
			return total, skipped, err
		}

		rows := batch.Elem()
		n := rows.Len()

		if n == 0 {
			return total, skipped, nil
		}

		values := make([][]any, n)

		for i := 0; i < n; i++ {
			rv := rows.Index(i).Elem()
			values[i] = make([]any, len(fields))

			for j, f := range fields {
				v, _ := f.ValueOf(write.Statement.Context, rv)
				values[i][j] = transferValue(v)
			}
		}

		first := total + skipped + 1

		if err = write.Exec("INSERT INTO ? ?", clause.Table{Name: name}, clause.Values{Columns: columns, Values: values}).Error; err == nil {
			total += int64(n)
		} else if !TransferSkipDuplicates[name] || !isDuplicateKeyError(err) {
			return total, skipped, fmt.Errorf("%s (rows %d to %d)", err, first, first+int64(n)-1)
		} else {
			// Inserts the rows one by one, skipping those whose key is already in the target.
			for i := range values {
				if rowErr := write.Exec("INSERT INTO ? ?", clause.Table{Name: name}, clause.Values{Columns: columns, Values: values[i : i+1]}).Error; rowErr == nil {
					total++
				} else if isDuplicateKeyError(rowErr) {
					skipped++
				} else {
					return total, skipped, fmt.Errorf("%s (row %d)", rowErr, first+int64(i))
				}
			}

			err = nil
		}

		if n < batchSize {
			return total, skipped, nil
		}

		lastRow := rows.Index(n - 1).Elem()
		last = make([]any, len(s.PrimaryFields))

		for i, f := range s.PrimaryFields {
			last[i], _ = f.ValueOf(write.Statement.Context, lastRow)
		}
	}
}

// transferKeyset returns the condition that selects the rows after the primary key values of the last row,
// e.g. "(a > ?)" or "(a >= ?) AND ((a > ?) OR (a = ? AND b > ?))". The leading bound of a composite key lets
// the databases start the index scan at the last row.
func transferKeyset(keys []*schema.Field, last []any) (string, []any) {
	// Without a value for each key, no row is selected, so that the count check reports the rows not copied.
	if len(keys) == 0 || len(last) < len(keys) {
		return "1 = 0", nil
	}

	last = last[:len(keys)]
	terms := make([]string, len(keys))
	args := make([]any, 0, len(keys)*(len(keys)+1)+2)

	if len(keys) > 1 {
		args = append(args, clause.Column{Name: keys[0].DBName}, last[0])
	}

	for i := range keys {
		parts := make([]string, 0, i+1)

		for j := 0; j < i; j++ {
			parts = append(parts, "? = ?")
			args = append(args, clause.Column{Name: keys[j].DBName}, last[j]) //nolint:gosec // last has a value for each key
		}

		parts = append(parts, "? > ?")
		args = append(args, clause.Column{Name: keys[i].DBName}, last[i]) //nolint:gosec // last has a value for each key
		terms[i] = "(" + strings.Join(parts, " AND ") + ")"
	}

	if len(keys) > 1 {
		return "(? >= ?) AND (" + strings.Join(terms, " OR ") + ")", args
	}

	return terms[0], args
}

// transferValue returns the value to bind for a field value. A byte slice of a named type, e.g. json.RawMessage,
// is bound as []byte, since GORM would write an empty one as NULL.
func transferValue(v any) any {
	if _, ok := v.(driver.Valuer); ok {
		return v
	} else if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice && rv.Type() != bytesType && rv.Type().Elem().Kind() == reflect.Uint8 {
		return rv.Convert(bytesType).Interface()
	}

	return v
}

// bytesType is the type of an unnamed byte slice.
var bytesType = reflect.TypeOf([]byte(nil))

// SameDatabase reports whether a and b are connected to the same database, by checking if a table that is
// created through b exists in a. It returns false without a check if the databases are of different kinds.
func SameDatabase(a, b *gorm.DB) (bool, error) {
	if a == nil || b == nil {
		return false, errors.New("database not connected")
	} else if dsn.DialectFromDriver(a.Name()) != dsn.DialectFromDriver(b.Name()) {
		return false, nil
	}

	name := "transfer_probe_" + strings.ToLower(rnd.Base36(12))

	if err := b.Exec("CREATE TABLE ? (id INTEGER)", clause.Table{Name: name}).Error; err != nil {
		return false, fmt.Errorf("%s (create %s)", err, clean.Log(name))
	}

	same := a.Migrator().HasTable(name)

	if err := b.Exec("DROP TABLE ?", clause.Table{Name: name}).Error; err != nil {
		return same, fmt.Errorf("%s (drop %s)", err, clean.Log(name))
	}

	return same, nil
}

// OtherConnections returns the number of connections to the database of db that do not belong to its own pool,
// e.g. of a PhotoPrism instance that still runs. It returns 0 for SQLite, which keeps no list of connections.
func OtherConnections(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("database not connected")
	}

	var query string

	switch dsn.DialectFromDriver(db.Name()) {
	case dsn.DialectMySQL:
		query = "SELECT COUNT(*) FROM information_schema.PROCESSLIST WHERE DB = DATABASE() AND ID <> CONNECTION_ID()"
	case dsn.DialectPostgreSQL:
		query = "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database() AND pid <> pg_backend_pid() AND backend_type = 'client backend'"
	default:
		return 0, nil
	}

	sqlDb, err := db.DB()

	if err != nil {
		return 0, err
	}

	var count int64

	if err = db.Raw(query).Row().Scan(&count); err != nil {
		return 0, err
	}

	// The connection that runs the query is excluded above, the other open connections of the pool here.
	if own := int64(sqlDb.Stats().OpenConnections - 1); own > 0 {
		count -= own
	}

	return max(count, 0), nil
}

// UserNameCollisions returns the number of accounts whose user name equals that of another account under the
// comparison rules of the database, e.g. names that differ only by accents on MariaDB.
func UserNameCollisions(db *gorm.DB) (count int64, err error) {
	if db == nil {
		return 0, errors.New("database not connected")
	}

	groups := db.Table(User{}.TableName()).Select("COUNT(*) AS n").
		Where("user_name <> ''").Group("user_name").Having("COUNT(*) > 1")

	err = db.Table("(?) AS collisions", groups).Select("COALESCE(SUM(n), 0)").Row().Scan(&count)

	return count, err
}
