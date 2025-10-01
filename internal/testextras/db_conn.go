package testextras

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Supported test databases.
const (
	MySQL           = "mysql"
	Postgres        = "postgres"
	SQLite3         = "sqlite"
	SQLiteTestDB    = ".test.db"
	SQLiteMemoryDSN = ":memory:?cache=shared&_foreign_keys=on"
	SQLiteMutexDSN  = "/go/src/github.com/photoprism/photoprism/storage/testdata/unit.mutex.db?_busy_timeout=5000&_foreign_keys=on"
)

var drivers = map[string]func(string) gorm.Dialector{
	MySQL:    mysql.Open,
	Postgres: postgres.Open,
	SQLite3:  sqlite.Open,
}

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

// Replicated here to avoid import loop.
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
			log,
			logger.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  logger.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      true,          // Don't include params in the SQL log
				Colorful:                  false,         // Disable color
			},
		),
		// Set UTC as the default for created and updated timestamps.
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
}
