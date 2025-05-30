package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fatkulllin/metrilo/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Database struct {
	dsn string
	db  *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {

	db, err := sql.Open("pgx", dsn)
	fmt.Println(dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &Database{db: db, dsn: dsn}, nil
}

func (d *Database) ReconnectDB() error {
	if d.db != nil {
		if err := d.Close(); err != nil {
			logger.Log.Error("Error closing DB connection", zap.Error(err))
		}
	}
	newDB, err := NewDatabase(d.dsn)
	if err != nil {
		return err
	}
	d.db = newDB.db
	return nil
}

func (d *Database) GetDB() *sql.DB {
	if d.db == nil {
		logger.Log.Error("Error database is not connected")
		return nil
	}
	return d.db
}

func (d *Database) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()

}
