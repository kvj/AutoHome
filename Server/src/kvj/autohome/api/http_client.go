package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kvj/autohome/data"
	"kvj/autohome/internet"
	_ "log"
	"math"
	"net/http"
)

type sensorDef [4]int

var sensors = map[string]sensorDef{
	"room_t":       {0, 0, 0, 1},
	"room_0_move":  {0, 1, 0, 0},
	"room_0_light": {0, 2, 0, 0},
	"room_1_move":  {1, 1, 0, 0},
	"room_1_light": {1, 2, 0, 0},
	"room_2_move":  {2, 1, 0, 0},
	"room_2_light": {2, 2, 0, 0},
	"room_3_move":  {3, 1, 0, 0},
	"room_3_light": {3, 2, 0, 0},
	"out_cond":     {10, 20, 0, 0},
	"out_t":        {10, 20, 0, 1},
	"out_wind":     {10, 20, 0, 5},
	"out_wsp":      {10, 20, 0, 6},
	"batt_0":       {30, 3, 0, 0},
	"batt_1":       {35, 3, 0, 0},
}

func url(conf data.HashMap, uri string) string {
	return fmt.Sprintf("%s%s?key=%s", conf["server"], uri, conf["key"])
}

func loadLatest(conf data.HashMap) (map[string]float64, error) {
	req := &appSensors{
		Actual:  true,
		Sensors: make([]appSensor, len(sensors)),
	}
	i := 0
	keys := make([]string, len(sensors))
	for key, arr := range sensors {
		req.Sensors[i] = appSensor{
			Device:  arr[0],
			Type:    arr[1],
			Index:   arr[2],
			Measure: arr[3],
		}
		keys[i] = key
		i++
	}
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(url(conf, "api/latest"), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64)
	for idx, key := range keys {
		result[key] = req.Sensors[idx].Value
	}
	return result, err
}

const bar = "_\\|/"

func makeBar(val float64, max float64) string {
	idx := int(math.Floor(val / max * float64(len(bar)+1)))
	return bar[idx : idx+1]
}

func makeFlag(val float64) string {
	if val == 0 {
		return "*"
	}
	return "#"
}

func MakeStatus(conf data.HashMap) (string, string) {
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: cfg,
	}

	vals, err := loadLatest(conf)
	if err != nil {
		// log.Printf("Error:", err)
		return "", "HTTP error"
	}
	outp := fmt.Sprintf("%s%s %s%.f : %s : %s%s%s%s%s%s%s%s : %.f%% : %.f%%",
		internet.Code2Char(vals["out_cond"]),
		internet.FormatTemp(vals["out_t"]),
		internet.FormatWindDirection(vals["out_wind"]),
		vals["out_wsp"],
		internet.FormatTemp(vals["room_t"]),
		makeFlag(vals["room_0_move"]),
		makeBar(vals["room_0_light"], 256),
		makeFlag(vals["room_1_move"]),
		makeBar(vals["room_1_light"], 256),
		makeFlag(vals["room_2_move"]),
		makeBar(vals["room_2_light"], 256),
		makeFlag(vals["room_3_move"]),
		makeBar(vals["room_3_light"], 256),
		vals["batt_0"],
		vals["batt_1"])
	// log.Printf("Data: %v %s", vals, outp)
	return outp, ""
}
