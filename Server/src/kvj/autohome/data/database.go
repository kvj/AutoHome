package data

import (
	"database/sql"
	_ "github.com/lib/pq"
	"kvj/autohome/model"
	"log"
	"time"
)

type DBProvider struct {
	db *sql.DB
}

func OpenDB(url string) (*DBProvider, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Printf("DB open error: %v", err)
		return nil, err
	}
	provider := &DBProvider{
		db: db,
	}
	return provider, nil
}

func (self *DBProvider) AddMeasure(device int, measure *model.MeasureMessage) error {
	tx, err := self.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("insert into measure "+
		"(device, type, sensor, measure, value, at) values "+
		"($1, $2, $3, $4, $5, $6)",
		device, measure.Type, measure.Sensor, measure.Measure, measure.Value, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
