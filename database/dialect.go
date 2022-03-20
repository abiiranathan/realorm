package database

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DialectString string

const (
	PG      DialectString = "postgres"
	SQLITE3 DialectString = "sqlite3"
	MYSQL   DialectString = "mysql"
)

// Constructs a DSN string from a config object c
func (c *config) getDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s", c.host, c.user, c.password, c.database, c.sslmode, c.timezone)
}

// Constructs a DSN string from a config object
func DSNFromConfig(c *config) string {
	return c.getDSN()
}

func parse_dsn(dsn string) map[string]string {
	// Parse DSN and extract connection parameters
	params := map[string]string{}
	for _, s := range strings.Split(dsn, " ") {
		v := strings.Split(s, "=")

		if len(v) == 2 {
			params[v[0]] = v[1]
		}
	}
	return params
}

// ParseDSN parses the DSN string to a config object
// Default value for sslmode is disable
// Default value for timezone is UTC
func ParseDSN(dsn string) (*config, error) {
	var c config

	if dsn == "" {
		return nil, fmt.Errorf("cannot parse DSN. %s", "DSN is empty")
	}

	params := parse_dsn(dsn)

	if sslmode, ok := params["sslmode"]; ok {
		c.sslmode = sslmode
	} else {
		c.sslmode = "disable"
	}

	if timezone, ok := params["timezone"]; ok {
		c.timezone = timezone
	} else {
		c.timezone = "UTC"
	}

	if host, ok := params["host"]; ok {
		c.host = host
	} else {
		return nil, fmt.Errorf("cannot parse DSN. %s", "host is empty")
	}

	if user, ok := params["user"]; ok {
		c.user = user
	} else {
		return nil, fmt.Errorf("cannot parse DSN. %s", "user is empty")
	}

	if password, ok := params["password"]; ok {
		c.password = password
	} else {
		return nil, fmt.Errorf("cannot parse DSN. %s", "password is empty")
	}

	if database, ok := params["dbname"]; ok {
		c.database = database
	} else {
		return nil, fmt.Errorf("cannot parse DSN. %s", "database is empty")
	}

	return &c, nil
}

func ParseDialect(dsn string, dialect DialectString) (gorm.Dialector, gorm.Config, error) {
	var dialector gorm.Dialector
	gormConfig := gorm.Config{Logger: logger.Default.LogMode(logger.Info)}

	switch os.Getenv("SQL_LOG_LEVEL") {
	case "info":
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	case "error":
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	default:
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	switch dialect {
	case PG:
		dialector = postgres.Open(dsn)
		gormConfig = gorm.Config{
			// Cache prepared statements to speedup performance
			PrepareStmt:                              true,
			Logger:                                   gormConfig.Logger,
			DisableForeignKeyConstraintWhenMigrating: false,
		}
	case SQLITE3:
		dialector = sqlite.Open(dsn)
	case MYSQL:
		dialector = mysql.Open(dsn)
	default:
		return nil, gormConfig, fmt.Errorf("unknown dialect: %s", dialect)
	}

	return dialector, gormConfig, nil
}
