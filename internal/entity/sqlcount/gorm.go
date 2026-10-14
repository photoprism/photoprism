package sqlcount

import (
	"gorm.io/gorm"
)

// Provider is a database connection provider whose statements are recorded by Counter.
type Provider struct {
	DB      *gorm.DB
	Counter *Counter
}

// Db returns the counted connection.
func (p *Provider) Db() *gorm.DB {
	return p.DB
}

// Close closes the counted connection.
func (p *Provider) Close() error {
	sqlDb, err := p.DB.DB()

	if err != nil {
		return err
	}

	return sqlDb.Close()
}

// OpenFunc opens an ORM connection that uses the connection pool passed as argument.
type OpenFunc func(conn gorm.ConnPool) (*gorm.DB, error)

// OpenGorm returns a connection provider for the driver and data source name whose statements are
// recorded by its Counter. Callers that swap it in as the entity provider must restore the previous one.
func OpenGorm(driverName, dataSourceName string, open OpenFunc) (*Provider, error) {
	sqlDb, c, err := Open(driverName, dataSourceName)

	if err != nil {
		return nil, err
	}

	db, err := open(sqlDb)

	if err != nil {
		_ = sqlDb.Close()
		return nil, err
	}

	return &Provider{DB: db, Counter: c}, nil
}
