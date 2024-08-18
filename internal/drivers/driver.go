package drivers

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

// jese app.config mai sara data hai isme poora database hai

// holds database connection pool
type DB struct {
	SQL *sql.DB
}

// DBConn is pointer to DB
var DBConn = &DB{}

const maxOpenDbConnection = 10
const maxIdleDbConnection = 5
const maxLifetimeDbConnection = 5 * time.Minute

func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}
	d.SetMaxOpenConns(maxOpenDbConnection)
	d.SetMaxIdleConns(maxIdleDbConnection)
	d.SetConnMaxLifetime(maxLifetimeDbConnection)
	DBConn.SQL = d
	err = TestDB(d)
	if err != nil {
		fmt.Println("Error testing connection to database")
		panic(err)
	}
	return DBConn, nil

}
func TestDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db, nil
}
