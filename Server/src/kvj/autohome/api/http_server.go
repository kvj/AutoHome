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

func dataFolder(config data.HashMap) string {
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
	Device  int    `json:"device"`
	Type    int    `json:"type"`
	Index   int    `json:"index"`
	Measure int    `json:"measure"`
	Plugin  string `json:"plugin"`
	Extra   string `json:"extra"`
	Revert  bool   `json:"revert"`
}

type appLayout struct {
	Position []int       `json:"position"`
	Sensors  []appSensor `json:"sensors"`
}

type appConfig struct {
	Keys   []string    `json:"keys"`
	Layout []appLayout `json:"layout"`
}

func loadConfig(config data.HashMap) *appConfig {
	folder := dataFolder(config)
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

func confApiHandler(config data.HashMap, body interface{}) (interface{}, string) {
	return loadConfig(config), ""
}

type jsonFactory func() interface{}
type apiHandler func(config data.HashMap, body interface{}) (interface{}, string)
type httpHandler func(w http.ResponseWriter, r *http.Request)

func addApiCall(config data.HashMap, factory jsonFactory, handler apiHandler) httpHandler {
	conf := loadConfig(config)
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
		bodyOut, errStr := handler(config, body)
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

func setupStatic(conf data.HashMap) {
	dataPath := dataFolder(conf)
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
	http.HandleFunc("/api/config", addApiCall(conf, func() interface{} {
		return &jsonEmpty{}
	}, confApiHandler))
}

func StartServer(conf data.HashMap) {
	setupStatic(conf)
	panic(http.ListenAndServe(fmt.Sprintf(":%s", conf["port"]), nil))
	// log.Printf("HTTP server started")
}
