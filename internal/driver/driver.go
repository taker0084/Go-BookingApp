package driver

import (
	"database/sql"
	"time"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
)

//DB holds the database connection pool
type DB struct{
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

//ConnectSQL creates database pool for postgres
func ConnectSQL(dsn string) (*DB, error){
	d, err := NewDataBase(dsn)
	if err != nil{
		panic(err)
	}
	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = d

	err = TestDB(d)
	if err != nil{
		return nil, err
	}
	return dbConn, nil
}

//NewDatabase creates a new database for the application
func NewDataBase(dsn string) (*sql.DB, error){
	db, err := sql.Open("pgx", dsn)
	if err != nil{
		return nil, err
	}
	if err = db.Ping(); err != nil{
		return nil, err
	}
	return db, nil
}

//TestDB tries to ping database
func TestDB(d *sql.DB) error{
	err := d.Ping()
	if err != nil{
		return err
	}
	return nil
}