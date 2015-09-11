package data

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lib/pq"
	"kvj/autohome/model"
	"log"
	"math"
	"time"
)

const (
	TypeMeasure          = iota
	TypeForecast         = iota
	MaxRows      float64 = 100
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
			log.Printf("Listener error: %v", err.Error())
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
			case <-time.Tick(2 * time.Minute):
				pingErr := listener.Ping()
				if pingErr != nil {
					log.Printf("Ping error: %v", pingErr)
				}
				continue
			case n := <-listener.Notify:
				// log.Printf("From channel:", n.Channel, n.Extra)
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

func (self *DBProvider) NotifyMeasure(m *model.MeasureNotification) {
	bodyOutBytes, err := json.Marshal(m)
	if err == nil {
		json := string(bodyOutBytes)
		self.Notify("sensor", json)
		// self.SensorChannel <- json
	} else {
		log.Printf("JSON error: %v", err)
	}
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
	rows_count, err := self.db.Query("select count(*) from "+table_name+" where device=$1 and type=$2 and sensor=$3 and measure=$4 and at between $5 and $6", device, _type, index, measure, time.Unix(from/1000, 0), time.Unix(to/1000, 0))
	if err != nil {
		log.Printf("Failed to get count:", err)
		return nil, nil, err
	}
	defer rows_count.Close()
	var count = 0
	if rows_count.Next() {
		err = rows_count.Scan(&count)
		if err != nil {
			log.Printf("Failed to get count:", err)
			return nil, nil, err
		}
	} else {
		log.Printf("No results for Count SQL")
	}
	if count == 0 {
		count = 1
	}
	rows, err := self.db.Query("select value, at from "+table_name+" where device=$1 and type=$2 and sensor=$3 and measure=$4 and at between $5 and $6 order by id", device, _type, index, measure, time.Unix(from/1000, 0), time.Unix(to/1000, 0))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var values []float64
	var times []*time.Time
	var batch_count = 0
	var batch_value = 0.0
	var batch_size = int(math.Ceil(float64(count) / MaxRows))
	var last_time *time.Time = nil
	for rows.Next() {
		var value float64
		var time time.Time
		err = rows.Scan(&value, &time)
		if err != nil {
			return nil, nil, err
		}
		batch_count += 1
		batch_value += value
		last_time = &time
		if batch_count == batch_size {
			values = append(values, batch_value/float64(batch_count))
			times = append(times, &time)
			batch_count = 0
			batch_value = 0.0
		}
	}
	if batch_count > 0 {
		values = append(values, batch_value/float64(batch_count))
		times = append(times, last_time)
	}
	// log.Printf("Period", count, batch_size, batch_count, last_time, len(values))
	// log.Printf("Data found:", device, _type, index, measure, value, time)
	return values, times, nil // OK

}

func (self *DBProvider) LatestMeasure(device, _type, index, measure int) (float64, *time.Time, error) {
	rows, err := self.db.Query("select value, at from measure where device=$1 and type=$2 and sensor=$3 and measure=$4 order by id desc limit 1", device, _type, index, measure)
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

func (self *DBProvider) ClosestForecast(device, _type, index, measure int, from int64) (float64, *time.Time, error) {
	rows, err := self.db.Query("select value, at from forecast where device=$1 and type=$2 and sensor=$3 and measure=$4 and at<$5 order by at desc limit 1", device, _type, index, measure, time.Unix(from/1000, 0))
	if err != nil {
		return 0, nil, err
	}
	var value float64
	var time time.Time
	defer rows.Close()
	if !rows.Next() {
		log.Printf("No data found:", device, _type, index, measure)
		return -1, nil, nil // No data
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
	var cameraPortVar = flag.String("camera-port", "9101", "HTTP camera port")
	var cameraURLVar = flag.String("camera-url", "rtsp://%s:554/user=admin&password=&channel=1&stream=0.sdp", "HTTP camera port")
	var ffmpegVar = flag.String("ffmpeg", "ffmpeg", "Path to ffmpeg")
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
	result["cameraPort"] = *cameraPortVar
	result["cameraUrl"] = *cameraURLVar
	result["ffmpeg"] = *ffmpegVar
	return result
}
