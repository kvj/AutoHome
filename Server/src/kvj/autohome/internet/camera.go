package internet

import (
	"fmt"
	// "io/ioutil"
	"kvj/autohome/data"
	// "kvj/autohome/model"
	"log"
	"net/http"
	// "strconv"
	// "time"
	"os/exec"
)

var (
	config data.HashMap
)

func escamSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	cameraHost := r.FormValue("host")
	cameraPass := r.FormValue("password")
	if cameraHost == "" {
		log.Printf("No camera host specified: %v", r.Form)
		http.Error(w, "Request error", 500)
		return
	}
	cameraUrl := fmt.Sprintf("rtsp://%s:554/user=admin&password=%s&channel=1&stream=0.sdp", cameraHost, cameraPass)
	cmd := exec.Command(config["ffmpeg"], "-i", cameraUrl, "-vframes", "1", "-qscale", "15", "-f", "image2", "-")
	bytes, err := cmd.Output()
	if err != nil {
		log.Printf("ffmpeg error: %v, %s", err, cameraUrl)
		http.Error(w, "Camera Error", 500)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	// log.Printf("Bytes: %v", len(bytes))
	w.Write(bytes)
}

func StartCameraProxy(conf data.HashMap) {
	config = conf
	log.Printf("HTTP camera server started")
	http.HandleFunc("/camera/escam", escamSnapshot)
	panic(http.ListenAndServe(fmt.Sprintf(":%s", config["cameraPort"]), nil))
}
