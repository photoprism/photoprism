package entity

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// dbConn is the global gorm.DB connection provider.
var dbConn Gorm

// Gorm is a gorm.DB connection provider interface.
type Gorm interface {
	Db() *gorm.DB
}

// DbConn is a gorm.DB connection provider.
type DbConn struct {
	Driver string
	Dsn    string

	once sync.Once
	db   *gorm.DB
	pool *pgxpool.Pool
}

// Db returns the gorm db connection.
func (g *DbConn) Db() *gorm.DB {
	g.once.Do(g.Open)

	if g.db == nil {
		log.Fatal("migrate: database not connected")
	}

	return g.db
}

func OpenPostgreSQL(dsn string) (db *sql.DB, pool *pgxpool.Pool) {
	ctx := context.Background()
	pgxPoolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	pgxPoolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
		})

		return nil
	}

	pool, err = pgxpool.NewWithConfig(ctx, pgxPoolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatal(err)
	}

	return stdlib.OpenDBFromPool(pool), pool
}

// Open creates a new gorm db connection.
func (g *DbConn) Open() {
	log.Infof("Opening DB connection with driver %s", g.Driver)
	var db *gorm.DB
	var err error
	if g.Driver == Postgres {
		postgresDB, pgxPool := OpenPostgreSQL(g.Dsn)
		g.pool = pgxPool
		db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
	} else {
		db, err = gorm.Open(drivers[g.Driver](g.Dsn), gormConfig())
	}

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			fmt.Printf("gorm.Open(%s, %s) %d\n", g.Driver, g.Dsn, i)
			if g.Driver == Postgres {
				postgresDB, pgxPool := OpenPostgreSQL(g.Dsn)
				g.pool = pgxPool
				db, err = gorm.Open(postgres.New(postgres.Config{Conn: postgresDB}), gormConfig())
			} else {
				db, err = gorm.Open(drivers[g.Driver](g.Dsn), gormConfig())
			}

			if db != nil && err == nil {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}

		if err != nil || db == nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	}
	log.Info("DB connection established successfully")

	if g.Driver != Postgres {
		sqlDB, _ := db.DB()

		sqlDB.SetMaxIdleConns(4)   // in config_db it uses c.DatabaseConnsIdle(), but we don't have the c here.
		sqlDB.SetMaxOpenConns(256) // in config_db it uses c.DatabaseConns(), but we don't have the c here.
	}

	g.db = db
}

// Close closes the gorm db connection.
func (g *DbConn) Close() {
	if g.db != nil {
		sqlDB, _ := g.db.DB()
		if err := sqlDB.Close(); err != nil {
			log.Fatal(err)
		}

		g.db = nil
	}
	if g.pool != nil {
		g.pool.Close()
	}
}

// IsDialect returns true if the given sql dialect is used.
func IsDialect(name string) bool {
	return name == Db().Dialector.Name()
}

// DbDialect returns the sql dialect name.
func DbDialect() string {
	return Db().Dialector.Name()
}

// SetDbProvider sets the Gorm database connection provider.
func SetDbProvider(conn Gorm) {
	dbConn = conn
}

// HasDbProvider returns true if a db provider exists.
func HasDbProvider() bool {
	return dbConn != nil
}

func gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.New(
			log, // This should be dummy.NewLogger(), to match GORM1.  Set to log before release...
			logger.Config{
				SlowThreshold:             time.Second,  // Slow SQL threshold
				LogLevel:                  logger.Error, // Log level  <-- This should be Silent to match GORM1, set to Error before release...
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      false,        // Don't include params in the SQL log
				Colorful:                  false,        // Disable color
			},
		),
		// Set UTC as the default for created and updated timestamps.
		NowFunc: func() time.Time {
			return UTC()
		},
	}
}
