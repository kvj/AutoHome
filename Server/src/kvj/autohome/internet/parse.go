package internet

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"kvj/autohome/data"
	"kvj/autohome/model"
	"log"
	"net/http"
	"time"
)

const (
	applicationID = "3sXlXNwO7mRQQWRmGXK5coX38omFWHU8HMqe7kcE"
	pushFile      = "push.json"
	pushRest      = 10
	pushTick      = 1
)

type pushMessage struct {
	Data     interface{} `json:"data"`
	Channels []string    `json:"channels"`
}

type outputMessage struct {
	Result bool   `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type pushConfig struct {
	Id   string `json:"id,omitempty"`
	Last string `json:"last,omitempty"`
}

func parseCall(body_in, body_out interface{}, uri, apiKey string) error {
	messageOutBytes, err := json.Marshal(body_in)
	if err != nil {
		log.Printf("Failed to serialize push message", err)
		return err
	}
	// log.Printf("About to send message:", message)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.parse.com/"+uri, bytes.NewReader(messageOutBytes))
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
	err = json.Unmarshal(bodyBytes, &body_out)
	// log.Printf("Push response: %v", obj)
	if err != nil {
		log.Printf("Response parse error", err)
		return err
	}
	return nil
}

type instanceRequest struct {
	InstallationId string `json:"installationId"`
	DeviceType     string `json:"deviceType"`
}

type instanceResponse struct {
}

func makeInstance(apiKey string) (string, error) {
	id := uuid.NewV4().String()
	req := &instanceRequest{
		InstallationId: id,
		DeviceType:     "embedded",
	}
	res := &instanceResponse{}
	err := parseCall(req, res, "1/installations/", apiKey)
	if err != nil {
		log.Printf("Failed to make installation: %s %v", id, err)
		return "", err
	}
	return id, nil
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

type pushRequest struct {
	InstallationId string `json:"installation_id"`
	OauthKey       string `json:"oauth_key"`
	Version        string `json:"v"`
	Last           string `json:"last,omitempty"`
}

type incomingMessage struct {
	Data json.RawMessage `json:"data"`
	Time string          `json:"time"`
}

func StartPushListener(folder string, apiKey string) model.JsonChannel {
	conf := &pushConfig{}
	err := data.ReadJsonFromFile(folder, pushFile, conf)
	if err != nil {
		// Read failed -> try to make new install ID
		err = data.WriteJsonToFile(folder, pushFile, conf) // Check can write or not
		if err != nil {
			log.Fatal("Config folder not writable")
		}
		log.Printf("Let's make new install ID")
		id, err := makeInstance(apiKey)
		if err != nil {
			log.Fatal("Failed to make installation")
		}
		conf.Id = id
		_ = data.WriteJsonToFile(folder, pushFile, conf) // Save
	} else {
		log.Printf("No action needed, load OK")
	}
	tick := time.Tick(pushTick * time.Minute)
	data_chan := make(chan string)
	json_chan := make(model.JsonChannel)
	var conn *tls.Conn
	go func() {
		for {
			select {
			case <-tick:
				if conn != nil && conn.ConnectionState().HandshakeComplete {
					_, err = conn.Write([]byte("{}\n"))
					if err != nil {
						log.Printf("Tick write failed: %v", err)
					}
				}
			case str := <-data_chan:
				log.Printf("Parse: %s", str)
				message := &incomingMessage{}
				err = json.Unmarshal([]byte(str), message)
				if err != nil {
					log.Printf("Message parse failed: %s %v", str, err)
					continue
				}
				conf.Last = message.Time
				_ = data.WriteJsonToFile(folder, pushFile, conf) // Save
				json_chan <- message.Data
			}
		}
	}()
	go func() {
		for {
			if conn != nil {
				conn.Close()
			}
			log.Printf("Sleeping now")
			time.Sleep(pushRest * time.Second)
			log.Printf("Will dial push.parse.com...")
			conn, err = tls.Dial("tcp", "push.parse.com:443", nil)
			if err != nil {
				log.Printf("Failed to open push link: %v", err)
				continue
			}
			req := &pushRequest{
				InstallationId: conf.Id,
				OauthKey:       applicationID,
				Version:        "e1.0.0",
				Last:           conf.Last,
			}
			bytes, _ := json.Marshal(req)
			message_str := string(bytes)
			// log.Printf("Init with: %s", message_str)
			_, err = conn.Write([]byte(message_str + "\n"))
			if err != nil {
				log.Printf("Failed to init: %v", err)
				continue
			}
			reader := bufio.NewReader(io.Reader(conn))
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Connection closed: %v", err)
					break
				}
				if line == "" || line == "{}\n" {
					// Tick
					continue
				}
				data_chan <- line
			}
		}
	}()
	return json_chan
}
