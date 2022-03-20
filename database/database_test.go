package database

import (
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"
)

type Entity struct {
	ID int
}

type TestSuite struct {
	connection any
	dialect    DialectString
	shouldFail bool
	logLevel   string
}

func TestNew(t *testing.T) {
	os.Setenv("SQL_LOG_LEVEL", "silent")

	tests := []TestSuite{
		{
			connection: "dbname=realorm user=realorm password=password host=localhost sslmode=disable TimeZone=Africa/Kampala",
			dialect:    PG,
		},
		{
			connection: NewConfig("realorm", "realorm", "password", "localhost", "disable", "Africa/Kampala"),
			dialect:    PG,
		},
		{
			connection: NewConfig("realorm", "realorm", "password", "localhost", "disable", "Africa/Kampala"),
			dialect:    PG,
		},
		{
			connection: NewConfig("", "", "", "localhost", "disable", "Africa/Kampala"),
			dialect:    SQLITE3,
			shouldFail: true,
		},
		{
			connection: "nabiizy:password@tcp(localhost:3306)/realorm?charset=utf8mb4&parseTime=True&loc=Local",
			dialect:    MYSQL,
		},
		{
			connection: SQLITE3_MEMORY_DB,
			dialect:    SQLITE3,
		},
		{
			connection: nil,
			dialect:    PG,
			shouldFail: true,
		},
		{
			connection: "dbname=realorm user=realorm password=password host=localhost sslmode=disable TimeZone=Africa/Kampala",
			dialect:    "wrong dialect",
			shouldFail: true,
		},
	}

	for _, test := range tests {
		conn, err := Connect(test.connection, test.dialect)

		if err != nil && !test.shouldFail {
			t.Errorf("error creating database: %v on %s on\n", err, test.dialect)
			return
		}

		if test.shouldFail && (err == nil || conn != nil) {
			t.Errorf("expected error creating database: got %v on %s\n", err, test.dialect)
			return
		}

		func(db *gorm.DB) {
			// If connection failed, db will be nil
			if db == nil {
				return
			}

			// auto migrate entity
			err := db.AutoMigrate(&Entity{})
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				t.Errorf("error auto migrating entity on dialect %v: %v\n ", test.dialect, err)
			}

		}(conn)
	}

	// test nil connection
	_, err := Connect(nil, PG)
	if err == nil {
		t.Errorf("expected error because of nil connection\n")
	}

}

func TestParseDialect(t *testing.T) {
	tests := []TestSuite{
		{
			connection: "dbname=realorm user=realorm password=password host=localhost sslmode=disable TimeZone=Africa/Kampala",
			dialect:    PG,
			logLevel:   "silent",
		},
		{
			connection: "nabiizy:password@tcp(localhost:3306)/realorm?charset=utf8mb4&parseTime=True&loc=Local",
			dialect:    MYSQL,
			logLevel:   "info",
		},
		{
			connection: SQLITE3_MEMORY_DB,
			dialect:    SQLITE3,
			logLevel:   "error",
		},
	}

	for _, test := range tests {
		os.Setenv("SQL_LOG_LEVEL", test.logLevel)
		_, _, err := ParseDialect(test.connection.(string), test.dialect)

		if err != nil {
			t.Errorf("error parsing dialect: %v\n", err)
			return
		}
	}
}

// Test wrong dialect string
func TestParseWrongDialect(t *testing.T) {
	os.Setenv("SQL_LOG_LEVEL", "silent")
	_, _, err := ParseDialect("", "wrong")

	if err == nil {
		t.Errorf("expected error parsing wrong dialect\n")
		return
	}
}

// TestParseDSN
func TestParseDSN(t *testing.T) {
	os.Setenv("SQL_LOG_LEVEL", "silent")
	config, err := ParseDSN("dbname=realorm user=realorm password=password host=localhost sslmode=disable timezone=Africa/Kampala")

	if err != nil {
		t.Errorf("error parsing dsn: %v\n", err)
		return
	}

	// test default values
	if config.host != "localhost" {
		t.Errorf("expected host to be localhost, got %s\n", config.host)
	}

	// user
	if config.user != "realorm" {
		t.Errorf("expected user to be realorm, got %s\n", config.user)
	}

	// password
	if config.password != "password" {
		t.Errorf("expected password to be password, got %s\n", config.password)
	}

	// database
	if config.database != "realorm" {
		t.Errorf("expected database to be realorm, got %s\n", config.database)
	}

	// sslmode
	if config.sslmode != "disable" {
		t.Errorf("expected sslmode to be disable, got %s\n", config.sslmode)
	}

	// timezone
	if config.timezone != "Africa/Kampala" {
		t.Errorf("expected timezone to be Africa/Kampala, got %s\n", config.timezone)
	}

	// test empty config
	config, err = ParseDSN("")
	if err == nil {
		// should fail
		t.Errorf("expected error parsing empty dsn\n")
	}

	// test config with one param for all params
	params := []string{"dbname=realorm", "user=realorm", "password=password", "host=localhost", "sslmode=disable", "timezone=Africa/Kampala", "password=password"}

	for _, param := range params {
		config, err = ParseDSN(param)
		if err == nil {
			// should fail
			t.Errorf("expected error parsing dsn: %v\n", err)
		}
	}

	// test dsn without password
	config, err = ParseDSN("dbname=realorm user=realorm host=localhost sslmode=disable timezone=Africa/Kampala")
	if err == nil {
		// should fail
		t.Errorf("expected error parsing dsn without password\n")
	}

	// test dsn without database
	config, err = ParseDSN("user=realorm password=password host=localhost sslmode=disable timezone=Africa/Kampala")
	if err == nil {
		// should fail
		t.Errorf("expected error parsing dsn without database\n")
	}
}
