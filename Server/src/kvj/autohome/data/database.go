package data

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/lib/pq"
	"kvj/autohome/model"
	"log"
	"time"
)

const (
	TypeMeasure  = iota
	TypeForecast = iota
)

type DBProvider struct {
	db             *sql.DB
	listener       *pq.Listener
	SensorChannel  chan string
	CommandChannel chan string
}

type HashMap map[string]string

func OpenDB(config HashMap) *DBProvider {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config["dbuser"],
		config["dbpass"],
		config["dbhost"],
		config["dbport"],
		config["db"])
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("DB open error: %v", err)
	}
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("Listener error: %v", err)
		}
	}
	listener := pq.NewListener(url, 10*time.Second, time.Hour, reportProblem)
	err = listener.Listen("sensor")
	if err != nil {
		log.Fatal("DB channel open error: %v", err)
	}
	err = listener.Listen("command")
	if err != nil {
		log.Fatal("DB channel2 open error: %v", err)
	}
	provider := &DBProvider{
		db:             db,
		listener:       listener,
		SensorChannel:  make(chan string),
		CommandChannel: make(chan string),
	}
	go func() {
		for {
			select {
			case n := <-listener.Notify:
				if n.Channel == "sensor" {
					provider.SensorChannel <- n.Extra
					continue
				}
				if n.Channel == "command" {
					provider.CommandChannel <- n.Extra
					continue
				}
				log.Printf("Unknown message: %v", n)
			}
		}
	}()
	return provider
}

func (self *DBProvider) Notify(channel string, payload string) {
	go func() {
		rows, err := self.db.Query("NOTIFY " + channel + ", '" + payload + "'")
		if err != nil {
			log.Printf("Failed to notify %v: %v", channel, err)
			return
		}
		rows.Close()
		// log.Printf("Notify OK")
	}()
}

func (self *DBProvider) DataForPeriod(table int, device, _type, index, measure int, from, to int64) ([]float64, []*time.Time, error) {
	table_name := "measure"
	if table == TypeForecast {
		table_name = "forecast"
	}
	rows, err := self.db.Query("select value, at from "+table_name+" where device=$1 and type=$2 and sensor=$3 and measure=$4 and at between $5 and $6 order by at", device, _type, index, measure, time.Unix(from/1000, 0), time.Unix(to/1000, 0))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var values []float64
	var times []*time.Time
	for rows.Next() {
		var value float64
		var time time.Time
		err = rows.Scan(&value, &time)
		if err != nil {
			return nil, nil, err
		}
		values = append(values, value)
		times = append(times, &time)
	}
	// log.Printf("Data found:", device, _type, index, measure, value, time)
	return values, times, nil // OK

}

func (self *DBProvider) LatestMeasure(device, _type, index, measure int) (float64, *time.Time, error) {
	rows, err := self.db.Query("select value, at from measure where device=$1 and type=$2 and sensor=$3 and measure=$4 order by at desc limit 1", device, _type, index, measure)
	if err != nil {
		return 0, nil, err
	}
	var value float64
	var time time.Time
	defer rows.Close()
	if !rows.Next() {
		log.Printf("No data found:", device, _type, index, measure)
		return value, &time, nil // No data
	}
	err = rows.Scan(&value, &time)
	// log.Printf("Data found:", device, _type, index, measure, value, time)
	return value, &time, err // Data found
}

func (self *DBProvider) DropForecast(device int, measure *model.MeasureMessage) error {
	tx, err := self.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("delete from forecast where device=$1 and type=$2 and sensor=$3",
		device, measure.Type, measure.Sensor)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (self *DBProvider) AddMeasure(device int, measure *model.MeasureMessage) error {
	measure.Time = time.Now()
	return self.AddMeasures(TypeMeasure, device, []*model.MeasureMessage{measure})
}

func (self *DBProvider) AddMeasures(table int, device int, measures []*model.MeasureMessage) error {
	tx, err := self.db.Begin()
	if err != nil {
		return err
	}
	table_name := "measure"
	if table == TypeForecast && len(measures) > 0 {
		table_name = "forecast"
		_, err := tx.Exec("delete from forecast where device=$1 and type=$2 and sensor=$3",
			device, measures[0].Type, measures[0].Sensor)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, measure := range measures {
		_, err = tx.Exec("insert into "+table_name+" "+
			"(device, type, sensor, measure, value, at) values "+
			"($1, $2, $3, $4, $5, $6)",
			device, measure.Type, measure.Sensor, measure.Measure, measure.Value, measure.Time)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	// log.Printf("Measures added: %v %s", len(measures), table_name)
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
