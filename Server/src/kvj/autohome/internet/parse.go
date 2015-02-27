package internet

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"kvj/autohome/model"
	"log"
	"net/http"
)

const (
	applicationID = "3sXlXNwO7mRQQWRmGXK5coX38omFWHU8HMqe7kcE"
)

type pushMessage struct {
	Data     interface{} `json:"data"`
	Channels []string    `json:"channels"`
}

type outputMessage struct {
	Result bool   `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func SendParsePush(body interface{}, apiKey string, channels []string) error {
	// bodyOutBytes, err := json.Marshal(body)
	// if err != nil {
	// 	log.Printf("Failed to serialize data part", err)
	// 	return err
	// }
	message := pushMessage{
		Channels: channels,
		Data:     body,
	}
	messageOutBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to serialize push message", err)
		return err
	}
	// log.Printf("About to send message:", message)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.parse.com/1/push", bytes.NewReader(messageOutBytes))
	req.Close = true
	if err != nil {
		log.Printf("Request not created", err)
		return err
	}
	req.Header.Add("Content-Type", "application/json; encoding=utf-8")
	req.Header.Add("X-Parse-Application-Id", applicationID)
	req.Header.Add("X-Parse-REST-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed", err)
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Request read error", err)
		return err
	}
	// log.Printf("Body: %s", body)
	obj := outputMessage{
		Result: false,
	}
	err = json.Unmarshal(bodyBytes, &obj)
	// log.Printf("Push response: %v", obj)
	if err != nil {
		log.Printf("Response parse error", err)
		return err
	}
	if !obj.Result {
		return model.NewStringError(obj.Error)
	}
	return nil
}
