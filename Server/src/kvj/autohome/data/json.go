package data

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func ReadJsonFromFile(folder, file string, obj interface{}) error {
	_file, err := os.Open(path.Join(folder, file))
	if err != nil {
		log.Printf("File not found: %s %s %v", folder, file, err)
		return err
	}
	defer _file.Close()
	body, err := ioutil.ReadAll(_file)
	if err != nil {
		log.Printf("File read error: %v", err)
		return err
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		log.Printf("Parse JSON failed: %v", err)
		return err
	}
	return nil
}

func WriteJsonToFile(folder, file string, obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Failed to marshal: %v", err)
		return err
	}
	err = ioutil.WriteFile(path.Join(folder, file), bytes, os.ModePerm)
	if err != nil {
		log.Printf("File write failed: ", folder, file, err)
		return err
	}
	return nil
}
