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
	Device    int     `json:"device"`
	Type      int     `json:"type"`
	Index     int     `json:"index"`
	Measure   int     `json:"measure"`
	Plugin    string  `json:"plugin"`
	Extra     string  `json:"extra"`
	Revert    bool    `json:"revert"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"ts"`
}

type appLayout struct {
	Position []int       `json:"position"`
	Sensors  []appSensor `json:"sensors"`
}

type appSensors struct {
	Sensors []appSensor `json:"sensors"`
}

type appConfig struct {
	Keys   []string    `json:"keys"`
	Layout []appLayout `json:"layout"`
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
		sensor.Timestamp = time.Unix()
	}
	return sensorsBody, ""
}

type jsonFactory func() interface{}
type apiHandler func(body interface{}) (interface{}, string)
type httpHandler func(w http.ResponseWriter, r *http.Request)

func addApiCall(factory jsonFactory, handler apiHandler) httpHandler {
	conf := loadConfig()
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Key")
		keyPresent := false
		for _, item := range conf.Keys {
			if item == key {
				keyPresent = true
				break
			}
		}
		if !keyPresent {
			log.Printf("Invalid key provided: %s %v", key, conf)
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
		log.Printf("Static: %s %s", fileName, ext)
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
	panic(http.ListenAndServe(fmt.Sprintf(":%s", config["port"]), nil))
	// log.Printf("HTTP server started")
}
