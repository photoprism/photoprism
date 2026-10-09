package sqlcount

import (
	"github.com/jinzhu/gorm"
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
	return p.DB.Close()
}

// OpenGorm returns a connection provider for the driver and data source name whose statements are
// recorded by its Counter. Callers that swap it in as the entity provider must restore the previous one.
func OpenGorm(driverName, dataSourceName string) (*Provider, error) {
	sqlDb, c, err := Open(driverName, dataSourceName)

	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(driverName, sqlDb)

	if err != nil {
		_ = sqlDb.Close()
		return nil, err
	}

	db.LogMode(false)

	return &Provider{DB: db, Counter: c}, nil
}
