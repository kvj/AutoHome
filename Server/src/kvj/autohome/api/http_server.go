package api

import (
	"fmt"
	"kvj/autohome/data"
	// "kvj/autohome/model"
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
		oneFile("static/index.html", w, r)
	})
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		oneFile(r.URL.Path[1:], w, r)
	})
}

func StartServer(conf data.HashMap) {
	setupStatic(conf)
	panic(http.ListenAndServe(fmt.Sprintf(":%s", conf["port"]), nil))
	// log.Printf("HTTP server started")
}
