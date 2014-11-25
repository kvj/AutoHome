package api

import (
	"fmt"
	"kvj/autohome/data"
	// "kvj/autohome/model"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	dbProvider *data.DBProvider
	config     data.HashMap
)

func dataFolder() string {
	dataPath, err := filepath.Abs(config["path"])
	if err != nil {
		log.Fatal("Data folder not found: %v", config["path"])
	}
	dataPath = path.Clean(dataPath)
	_, err = os.Stat(dataPath)
	if err != nil {
		log.Fatal("Data folder not exists %v: %v", dataPath, err)
	}
	return dataPath
}

var mimes = map[string]string{
	"html": "text/html; charset=utf-8",
	"js":   "application/javascript; charset=utf-8",
	"css":  "text/css; charset=utf-8",
	"png":  "image/png",
}

type jsonEmpty struct{}

type appSensor struct {
	Device    int             `json:"device"`
	Type      int             `json:"type"`
	Index     int             `json:"index"`
	Measure   int             `json:"measure"`
	Plugin    string          `json:"plugin"`
	Extra     string          `json:"extra"`
	Revert    bool            `json:"revert"`
	Value     float64         `json:"value"`
	Timestamp int64           `json:"ts"`
	From      int64           `json:"from,omitempty"`
	To        int64           `json:"to,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type appLayout struct {
	Position []int       `json:"position"`
	Sensors  []appSensor `json:"sensors"`
}

type appSensors struct {
	Sensors []appSensor `json:"sensors"`
}

type appSeriesRequest struct {
	Series   []appSensor `json:"series"`
	Forecast bool        `json:"forecast"`
}

type appSeriesResponse struct {
	Series [][]appSensor `json:"series"`
}

type appConfig struct {
	Keys   []string    `json:"keys"`
	Layout []appLayout `json:"layout"`
}

type pushMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func loadConfig() *appConfig {
	folder := dataFolder()
	file, err := os.Open(path.Join(folder, config["config"]))
	if err != nil {
		log.Fatal("Config file not found: %v", config["config"])
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("Config read error %v", err)
	}
	file.Close()
	obj := &appConfig{}
	err = json.Unmarshal(body, obj)
	if err != nil {
		log.Fatal("Parse failed: %v", err)
	}
	return obj
}

func confApiHandler(body interface{}) (interface{}, string) {
	return loadConfig(), ""
}

func dataApiHandler(body interface{}) (interface{}, string) {
	seriesBody, ok := body.(*appSeriesRequest)
	if !ok {
		return nil, "Input data error"
	}
	response := &appSeriesResponse{
		Series: make([][]appSensor, len(seriesBody.Series)),
	}
	dataType := data.TypeMeasure
	if seriesBody.Forecast {
		dataType = data.TypeForecast
	}
	for idx, sensor := range seriesBody.Series {
		values, times, err := dbProvider.DataForPeriod(dataType, sensor.Device, sensor.Type, sensor.Index, sensor.Measure, sensor.From, sensor.To)
		if err != nil {
			log.Printf("Failed to load data: %v", err)
			return nil, "DB error"
		}
		arr := make([]appSensor, len(values))
		for i, _ := range values {
			arr[i] = appSensor{
				Value:     values[i],
				Timestamp: times[i].Unix() * 1000,
			}
		}
		response.Series[idx] = arr
	}
	return response, ""
}

func latestApiHandler(body interface{}) (interface{}, string) {
	sensorsBody, ok := body.(*appSensors)
	if !ok {
		return nil, "Input data error"
	}
	for idx, _ := range sensorsBody.Sensors {
		sensor := &sensorsBody.Sensors[idx]
		value, time, err := dbProvider.LatestMeasure(sensor.Device, sensor.Type, sensor.Index, sensor.Measure)
		if err != nil {
			log.Printf("Failed to load data: %v", err)
			return nil, "DB error"
		}
		// log.Printf("Data loaded: %v %v", value, time.Unix())
		sensor.Value = value
		sensor.Timestamp = time.Unix() * 1000
	}
	return sensorsBody, ""
}

type jsonFactory func() interface{}
type apiHandler func(body interface{}) (interface{}, string)
type httpHandler func(w http.ResponseWriter, r *http.Request)

func checkKey(conf *appConfig, r *http.Request) (string, bool) {
	key := r.Header.Get("X-Key")
	if key == "" {
		// in query?
		key = r.URL.Query().Get("key")
	}
	keyPresent := false
	for _, item := range conf.Keys {
		if item == key {
			keyPresent = true
			break
		}
	}
	if !keyPresent {
		return "", false
	}
	return key, true
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	conf := loadConfig()
	key, valid := checkKey(conf, r)
	if !valid {
		log.Printf("Invalid key provided: %s %v", key, conf.Keys)
		http.Error(w, "Invalid Key", 401)
		return
	}
	log.Printf("Have new link: %v", key)
	w.Header().Set("Content-Type", "text/event-stream;charset=utf-8")
	w.WriteHeader(200)
	sendData := func(data string) error {
		_, err := w.Write([]byte("data:" + data + "\n\n"))
		if err != nil {
			// Close
			return err
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return nil
	}
	sendPushMessage := func(m string, mtype string) (string, error) {
		mout := &pushMessage{
			Type: mtype,
		}
		err := mout.Data.UnmarshalJSON([]byte(m))
		if err != nil {
			log.Printf("Failed to unmarshal %v", err)
			return "", err
		}
		bytes, err := json.Marshal(mout)
		if err != nil {
			log.Printf("Failed to marshal %v", err)
			return "", err
		}
		return string(bytes), nil
	}
	sendData("")
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case m := <-dbProvider.SensorChannel:
			// log.Printf("Have sensor data: %v", m)
			str, err := sendPushMessage(m, "sensor")
			if err == nil {
				sendData(str)
			}
		case <-ticker.C:
			// Ping connection
			err := sendData("")
			if err != nil {
				// Failed to write
				log.Printf("Failed to ping: %v", err)
				ticker.Stop()
				return
			}
		}
	}
}

func sseThread() {
	go func() {
	}()
}

func addApiCall(factory jsonFactory, handler apiHandler) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		conf := loadConfig()
		key, valid := checkKey(conf, r)
		if !valid {
			log.Printf("Invalid key provided: %s %v", key, conf.Keys)
			http.Error(w, "Invalid Key", 401)
			return
		}
		body := factory()
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Body read failed: %v", err)
			http.Error(w, "Request error", 500)
			return
		}
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			log.Printf("Body parsing failed: %v", err)
			http.Error(w, "Request error", 500)
			return
		}
		bodyOut, errStr := handler(body)
		if errStr != "" {
			log.Printf("Internal error: %s", errStr)
			http.Error(w, errStr, 403)
			return
		}
		bodyOutBytes, err := json.Marshal(bodyOut)
		if err != nil {
			log.Printf("Body sending failed: %v", err)
			http.Error(w, "Response error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(bodyOutBytes)
	}
}

func setupStatic() {
	dataPath := dataFolder()
	oneFile := func(path string, w http.ResponseWriter, r *http.Request) {
		fileName := fmt.Sprintf("%s/%s", dataPath, path)
		ext := filepath.Ext(fileName)[1:]
		// log.Printf("Static: %s %s", fileName, ext)
		if mime, found := mimes[ext]; found {
			w.Header().Set("Content-Type", mime)
		}
		http.ServeFile(w, r, fileName)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			oneFile("static/index.html", w, r)
			return
		}
		// Error
		log.Printf("Not supported URL: %v", r.URL.Path)
		http.NotFound(w, r)
	})
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		oneFile(r.URL.Path[1:], w, r)
	})
}

func StartServer(conf data.HashMap, db *data.DBProvider) {
	dbProvider = db
	config = conf
	setupStatic()
	http.HandleFunc("/api/config", addApiCall(func() interface{} {
		return &jsonEmpty{}
	}, confApiHandler))
	http.HandleFunc("/api/latest", addApiCall(func() interface{} {
		return &appSensors{}
	}, latestApiHandler))
	http.HandleFunc("/api/data", addApiCall(func() interface{} {
		return &appSeriesRequest{}
	}, dataApiHandler))
	http.HandleFunc("/api/link", sseHandler)
	sseThread()
	panic(http.ListenAndServe(fmt.Sprintf(":%s", config["port"]), nil))
	// log.Printf("HTTP server started")
}
