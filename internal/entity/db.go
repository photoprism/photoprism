package entity

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// associationCallbacks are the GORM callbacks that save related records with their parent.
var associationCallbacks = []string{"gorm:save_before_associations", "gorm:save_after_associations"}

// Db returns the default *gorm.DB connection.
func Db() *gorm.DB {
	if dbConn == nil {
		return nil
	}

	return dbConn.Db()
}

// UnscopedDb returns an unscoped *gorm.DB connection
// that returns all records including deleted records.
// Like Db, it can be the base of several queries, as each derives its own statement.
func UnscopedDb() *gorm.DB {
	return Db().Unscoped().Session(&gorm.Session{})
}

// GormConfig returns the ORM settings for new database connections.
// Timestamps are UTC truncated to whole seconds, as the schema stores them without fractions.
// Migrations ignore relationships, so that migrating a model never migrates the models it refers to.
func GormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:                                   logger.Discard,
		NowFunc:                                  Now,
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	}
}

// GormDialector returns the ORM dialector for a database driver and data source name.
// If conn is not nil, the dialector uses it instead of opening a connection pool.
func GormDialector(driver, dataSourceName string, conn gorm.ConnPool) (gorm.Dialector, error) {
	switch dsn.ParseDriver(driver) {
	case dsn.DriverMySQL:
		// Columns keep the types of existing schemas: VARCHAR(255) for strings without a size
		// and DATETIME without fractional seconds.
		return mysqlDialector{Dialector: mysql.New(mysql.Config{
			DSN:                      dataSourceName,
			Conn:                     conn,
			DefaultStringSize:        255,
			DisableDatetimePrecision: true,
		}).(*mysql.Dialector)}, nil
	case dsn.DriverSQLite3:
		return sqliteDialector{Dialector: &sqlite.Dialector{DSN: dataSourceName, Conn: conn}}, nil
	default:
		return nil, fmt.Errorf("unsupported database driver %s", driver)
	}
}

// OpenDb opens a database connection with the ORM settings PhotoPrism uses.
func OpenDb(driver, dataSourceName string) (*gorm.DB, error) {
	return OpenDbConn(driver, dataSourceName, nil)
}

// OpenDbConn opens a database connection like OpenDb, using conn as connection pool if it is not nil.
func OpenDbConn(driver, dataSourceName string, conn gorm.ConnPool) (*gorm.DB, error) {
	dialector, err := GormDialector(driver, dataSourceName, conn)

	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, GormConfig())

	if err != nil {
		return nil, err
	}

	if err = DisableAssociationSaves(db); err != nil {
		_ = CloseDb(db)
		return nil, err
	}

	db.ClauseBuilders[clause.Where{}.Name()] = buildWhere

	return db, nil
}

// buildWhere builds a WHERE clause with each condition in parentheses, so that a condition with
// OR operators does not combine with the conditions next to it. Conditions grouped by Not, Or or a
// nested Where are left as they are.
func buildWhere(c clause.Clause, builder clause.Builder) {
	if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
		exprs := make([]clause.Expression, len(where.Exprs))

		for i, e := range where.Exprs {
			switch expr := e.(type) {
			case clause.Expr:
				expr.SQL = parenthesize(expr.SQL)
				exprs[i] = expr
			case clause.NamedExpr:
				expr.SQL = parenthesize(expr.SQL)
				exprs[i] = expr
			default:
				exprs[i] = e
			}
		}

		c.Expression = clause.Where{Exprs: exprs}
	}

	c.Build(builder)
}

// parenthesize returns a SQL condition in parentheses. A leading placeholder is separated from the
// opening parenthesis, as the ORM would otherwise expand a slice value bound to it into a list.
func parenthesize(sql string) string {
	switch {
	case strings.TrimSpace(sql) == "":
		return sql
	case strings.HasPrefix(sql, "?"):
		return "( " + sql + ")"
	default:
		return "(" + sql + ")"
	}
}

// DisableAssociationSaves removes the callbacks that create and update related records,
// so that saving a model writes only its own row and related records are saved explicitly.
func DisableAssociationSaves(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	for _, name := range associationCallbacks {
		if err := db.Callback().Create().Remove(name); err != nil {
			return err
		}

		if err := db.Callback().Update().Remove(name); err != nil {
			return err
		}
	}

	if err := db.Callback().Update().Before("gorm:update").Register("photoprism:omit_blank_fields", omitBlankFields); err != nil {
		return err
	} else if err = db.Callback().Update().Before("gorm:update").Register("photoprism:bind_primary_keys", bindPrimaryKeys); err != nil {
		return err
	} else if err = db.Callback().Query().Before("gorm:query").Register("photoprism:bind_primary_keys", bindPrimaryKeys); err != nil {
		return err
	}

	return db.Callback().Update().After("gorm:update").Register("photoprism:save_fallback", saveFallback)
}

// isSave reports whether an update writes all fields of a model, as Save does.
func isSave(stmt *gorm.Statement) bool {
	return stmt.Schema != nil && len(stmt.Selects) == 1 && stmt.Selects[0] == "*" &&
		stmt.ReflectValue.Kind() == reflect.Struct
}

// omitBlankFields leaves a blank CreatedAt and blank foreign keys that have a default value out of a Save,
// so that saving a model without them keeps the stored values.
func omitBlankFields(db *gorm.DB) {
	stmt := db.Statement

	if db.Error != nil || !isSave(stmt) {
		return
	}

	if f := stmt.Schema.LookUpField("CreatedAt"); f != nil {
		if _, isZero := f.ValueOf(stmt.Context, stmt.ReflectValue); isZero {
			stmt.Omits = append(stmt.Omits, f.DBName)
		}
	}

	for _, rel := range stmt.Schema.Relationships.BelongsTo {
		for _, ref := range rel.References {
			if f := ref.ForeignKey; f != nil && f.HasDefaultValue {
				if _, isZero := f.ValueOf(stmt.Context, stmt.ReflectValue); isZero {
					stmt.Omits = append(stmt.Omits, f.DBName)
				}
			}
		}
	}
}

// bindPrimaryKeys adds the zero parts of a composite primary key to the conditions of a statement
// whose model has its first key part set, as the ORM would otherwise bind only the parts that are set.
func bindPrimaryKeys(db *gorm.DB) {
	stmt := db.Statement

	if db.Error != nil || stmt.Schema == nil || len(stmt.Schema.PrimaryFields) < 2 ||
		stmt.ReflectValue.Kind() != reflect.Struct || stmt.ReflectValue.Type() != stmt.Schema.ModelType {
		return
	}

	first := stmt.Schema.PrioritizedPrimaryField

	if first == nil {
		first = stmt.Schema.PrimaryFields[0]
	}

	if _, isZero := first.ValueOf(stmt.Context, stmt.ReflectValue); isZero {
		return
	}

	var zero []clause.Expression

	for _, f := range stmt.Schema.PrimaryFields {
		if value, isZero := f.ValueOf(stmt.Context, stmt.ReflectValue); isZero {
			zero = append(zero, clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: f.DBName}, Value: value})
		}
	}

	if len(zero) > 0 {
		stmt.AddClause(clause.Where{Exprs: zero})
	}
}

// saveFallback completes a Save whose update matched no changed row: it reports an existing row as
// one affected row, and creates a missing one with its hooks, so the ORM does not insert it again.
// Rows are matched by primary key only; conditions added to a Save are not applied.
func saveFallback(db *gorm.DB) {
	stmt := db.Statement

	if db.Error != nil || db.RowsAffected != 0 || db.DryRun || !isSave(stmt) || len(stmt.Schema.PrimaryFields) == 0 {
		return
	}

	query := db.Session(&gorm.Session{NewDB: true}).Model(stmt.Model)

	if stmt.Unscoped {
		query = query.Unscoped()
	}

	for _, f := range stmt.Schema.PrimaryFields {
		value, isZero := f.ValueOf(stmt.Context, stmt.ReflectValue)

		if isZero {
			return
		}

		query = query.Where(clause.Eq{Column: clause.Column{Name: f.DBName}, Value: value})
	}

	var found int64

	if err := query.Limit(1).Count(&found).Error; err != nil {
		_ = db.AddError(err)
		return
	} else if found > 0 {
		db.RowsAffected = 1
		return
	}

	res := db.Session(&gorm.Session{NewDB: true}).Create(stmt.Model)

	if res.Error != nil {
		_ = db.AddError(res.Error)
	} else {
		db.RowsAffected = res.RowsAffected
	}
}

// CloseDb closes the connection pool of a database connection.
func CloseDb(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDb, err := db.DB()

	if err != nil {
		return err
	}

	return sqlDb.Close()
}
