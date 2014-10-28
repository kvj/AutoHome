package data

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"kvj/autohome/model"
	"log"
	"time"
)

type DBProvider struct {
	db *sql.DB
}

type HashMap map[string]string

func OpenDB() (*DBProvider, error) {
	config := MakeConfig()
	if nil == config {
		log.Fatal("Invalid config")
	}
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config["dbuser"],
		config["dbpass"],
		config["dbhost"],
		config["dbport"],
		config["db"])
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

func MakeConfig() HashMap {
	var result = HashMap{}
	var dbVar = flag.String("db", "arduino", "DB Name")
	var dbhostVar = flag.String("dbhost", "localhost", "DB Hostname")
	var dbportVar = flag.String("dbport", "5432", "DB Port")
	var dbuserVar = flag.String("dbuser", "arduino", "DB Username")
	var dbpassVar = flag.String("dbpass", "arduino", "DB Password")
	var portVar = flag.String("port", "9100", "HTTP port")
	var pathVar = flag.String("path", "../Web", "Data folder")
	var fileVar = flag.String("file", "config.json", "Configuration file")
	flag.Parse()
	if !flag.Parsed() {
		flag.PrintDefaults()
		return nil
	}
	result["dbhost"] = *dbhostVar
	result["dbport"] = *dbportVar
	result["db"] = *dbVar
	result["dbuser"] = *dbuserVar
	result["dbpass"] = *dbpassVar
	result["port"] = *portVar
	result["path"] = *pathVar
	result["config"] = *fileVar
	return result
}
