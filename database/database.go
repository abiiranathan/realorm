package database

import (
	"fmt"

	"gorm.io/gorm"
)

const SQLITE3_MEMORY_DB = "file::memory:?cache=shared"

type config struct {
	database string
	user     string
	password string
	host     string
	sslmode  string
	timezone string
}

// Connect to the database using the config object or dsn
// string. It returns a pointer to the database connection
// and an error if any.
func Connect(connection any, dialect DialectString) (*gorm.DB, error) {
	var err error
	var dsn string

	switch connection.(type) {
	case *config:
		// config is valid only if dialect is postgres
		if dialect == PG {
			dsn = DSNFromConfig(connection.(*config))
		} else {
			return nil, fmt.Errorf("config is only valid when dialect is postgres")
		}
	case string:
		dsn = connection.(string)
	default:
		return nil, fmt.Errorf("connection is not a valid dsn string or config")
	}

	dialector, gormConfig, err := ParseDialect(dsn, dialect)

	if err != nil {
		return nil, err
	}

	return gorm.Open(dialector, &gormConfig)

}

// Returns a pointer to the new config object
func NewConfig(db, user, password, host, ssl_mode, tz string) *config {
	return &config{
		database: db,
		user:     user,
		password: password,
		host:     host,
		sslmode:  "disable",
		timezone: tz,
	}
}
