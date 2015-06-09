package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	// "path"
	// "path/filepath"
	// "strings"
	"time"
)

const (
	CHARS = "1234567890abcdef"
)

type CameraFile struct {
	Path, Code string
	Created    time.Time
}

var (
	cameraCache map[string]*CameraFile
)

func randomStr(size int) string {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = CHARS[rand.Int63n(int64(len(CHARS)))]
	}
	return string(b)
}

func downloadToTemp(url string) (*os.File, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to download: %v, %v", url, err)
		return nil, err
	}
	file, err := ioutil.TempFile("", "camera")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return nil, err
	}
	bytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Failed to read: %v, %v", url, err)
		return nil, err
	}
	_, err = file.Write(bytes)
	file.Close()
	return file, nil
}

func Download(server, data_type, host string) (string, string) {
	// Download snapshot from camera and save in temp file
	var url = ""
	switch data_type {
	case "escam_snapshot":
		url = fmt.Sprintf("http://%s/camera/escam?host=%s", server, host)
	default:
		return "", "Invalid type"
	}
	file, err := downloadToTemp(url)
	if err != nil {
		return "", "Download error"
	}
	rec := &CameraFile{
		Path:    file.Name(),
		Created: time.Now(),
		Code:    randomStr(12),
	}
	cameraCache[rec.Code] = rec
	// log.Printf("Downloaded: %s %s", rec.Path, rec.Code)
	return rec.Code, ""
}

func WriteSaved(code string, writer io.Writer) string {
	rec, ok := cameraCache[code]
	if !ok {
		return "Not found"
	}
	file, err := os.Open(rec.Path)
	if err != nil {
		log.Printf("File not found: %v", rec.Path)
		return "Not found"
	}
	defer file.Close()
	io.Copy(writer, file)
	return "" // Copied
}

func StartTempFileWatcher(seconds int, olderMin int) {
	cameraCache = make(map[string]*CameraFile)
	go func() {
		ticker := time.NewTicker(time.Duration(seconds) * time.Second)
		for {
			select {
			case <-ticker.C:
				// log.Printf("Cleanup old images")
				olderThan := time.Now().Add(-time.Duration(olderMin) * time.Minute)
				for code, rec := range cameraCache {
					if rec.Created.Before(olderThan) {
						delete(cameraCache, code)
						err := os.Remove(rec.Path)
						log.Printf("Removed old image: %v %s", err, rec.Path)
					}
				}
			}
		}
	}()
}
