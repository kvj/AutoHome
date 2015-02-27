package internet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kvj/autohome/data"
	"kvj/autohome/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	wundergroundURL  = "http://api.wunderground.com/api"
	wundergroundKey  = "36db8709d42795ff"
	wundergroundType = 20
)

// var forecastHours = [...]int{3, 6}
var forecastHours = [...]int{3, 6, 12, 24, 48}

type Crawler struct {
	queue    model.MMChannel
	forecast model.MMsChannel
	index    int
	ticker   *time.Ticker
}

func (self *Crawler) json(url string, obj interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Request failed %s %v", url, err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Request read error %s %v", url, err)
		return err
	}
	// log.Printf("Body: %s", body)
	err = json.Unmarshal(body, &obj)
	// log.Printf("Result: %v", obj)
	return err
}

type jsonConditionsResponse struct {
	Temp_c, Wind_degrees, Wind_kph                                         float64
	Icon, Relative_humidity, Precip_today_metric, Feelslike_c, Pressure_mb string
}

type jsonConditions struct {
	Current_observation jsonConditionsResponse
}

type jsonEnglishMetric struct {
	Metric string
}

type jsonFCTTIME struct {
	Epoch string
}

type jsonWindDir struct {
	Degrees string
}

type jsonHourlyForecast struct {
	Icon, Humidity, Pop, Sky               string
	Temp, Wspd, Feelslike, Qpf, Snow, Mslp jsonEnglishMetric
	FCTTIME                                jsonFCTTIME
	Wdir                                   jsonWindDir
}

type jsonHourly struct {
	Hourly_forecast []jsonHourlyForecast
}

func icon2Code(icon string) int {
	switch icon {
	case "chanceflurries", "chancesleet", "chancesnow", "sleet", "snow":
		return 6
	case "chancerain", "rain":
		return 4
	case "chancetstorms", "tstorms":
		return 5
	case "clear", "sunny":
		return 1
	case "cloudy", "mostlycloudy":
		return 3
	case "mostlysunny", "partlycloudy", "partlysunny":
		return 2
	case "fog", "hazy":
		return 7
	}
	return 0
}

func code2Text(code int) string {
	switch code {
	case 6:
		return "Snow"
	case 4:
		return "Rain"
	case 5:
		return "T-Storms"
	case 1:
		return "Clear"
	case 3:
		return "Cloudy"
	case 2:
		return "P. Cloudy"
	case 7:
		return "Fog"
	}
	return "???"
}

func str2Float(value string) float64 {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return f
}

type oneHourlyRecord struct {
	tstamp time.Time
	data   weatherMeasure
}

type WeatherCrawler struct {
	Crawler
	location string
	now      weatherMeasure
	hourly   []*oneHourlyRecord
}

type weatherMeasure struct {
	temp, wind, wspeed, weather, humidity, rain, snow, feels, pressure, probability, sky float64
}

func current2WeatherMeasure(current jsonConditionsResponse) weatherMeasure {
	var humidity float64
	if len(current.Relative_humidity) > 1 {
		humidity = str2Float(current.Relative_humidity[0 : len(current.Relative_humidity)-1])
	} else {
		// Invalid value
		humidity = 0
	}
	return weatherMeasure{
		temp:     current.Temp_c,
		wind:     current.Wind_degrees,
		wspeed:   current.Wind_kph,
		weather:  float64(icon2Code(current.Icon)),
		humidity: humidity,
		rain:     str2Float(current.Precip_today_metric),
		feels:    str2Float(current.Feelslike_c),
		pressure: str2Float(current.Pressure_mb),
	}
}

func hourly2WeatherMeasure(current jsonHourlyForecast) weatherMeasure {
	return weatherMeasure{
		temp:        str2Float(current.Temp.Metric),
		wind:        str2Float(current.Wdir.Degrees),
		wspeed:      str2Float(current.Wspd.Metric),
		weather:     float64(icon2Code(current.Icon)),
		humidity:    str2Float(current.Humidity),
		rain:        str2Float(current.Qpf.Metric),
		snow:        str2Float(current.Snow.Metric),
		probability: str2Float(current.Pop),
		feels:       str2Float(current.Feelslike.Metric),
		pressure:    str2Float(current.Mslp.Metric),
		sky:         str2Float(current.Sky),
	}
}

func (self *WeatherCrawler) putMeasureMessageToArray(arr model.MeasureMessages, from int, values []float64) {
	for idx, value := range values {
		arr[idx+from] = &model.MeasureMessage{
			Type:    wundergroundType,
			Sensor:  self.index,
			Measure: idx,
			Value:   value,
		}
	}
}

func (self *WeatherCrawler) poll() {
	// One call
	log.Printf("One call %v %v", self.location, len(self.hourly))
	url := fmt.Sprintf("%s/%s/conditions/q/%s.json", wundergroundURL, wundergroundKey, self.location)
	var json jsonConditions
	err := self.json(url, &json)
	if err != nil {
		log.Printf("Conditions error: %v", err)
		return
	}
	current := current2WeatherMeasure(json.Current_observation)
	log.Printf("Conditions: %+v", current)
	self.now = current
	// Make measurements
	// log.Printf("Message prepared:", message)
	values := []float64{current.weather, current.temp, current.feels, current.humidity, current.pressure, current.wind, current.wspeed, current.rain}
	messages := make(model.MeasureMessages, len(values))
	self.putMeasureMessageToArray(messages, 0, values)
	for _, message := range messages {
		self.queue <- message
	}

	var hourly jsonHourly
	url = fmt.Sprintf("%s/%s/hourly10day/q/%s.json", wundergroundURL, wundergroundKey, self.location)
	err = self.json(url, &hourly)
	if err != nil {
		log.Printf("Hourly error: %v", err)
		return
	}
	// log.Printf("Hourly: %+v", hourly)
	self.hourly = make([]*oneHourlyRecord, len(hourly.Hourly_forecast))
	for idx, item := range hourly.Hourly_forecast {
		timeSec := str2Float(item.FCTTIME.Epoch)
		timeObj := time.Unix(int64(timeSec), 0)
		data := hourly2WeatherMeasure(item)
		// log.Printf("Hour: %+v %v", data, timeObj)
		self.hourly[idx] = &oneHourlyRecord{
			tstamp: timeObj,
			data:   data,
		}
	}
	const values_size = 11
	arr := make(model.MeasureMessages, len(self.hourly)*values_size)
	from := 0
	for _, item := range self.hourly {
		item_values := []float64{item.data.weather, item.data.temp, item.data.feels, item.data.humidity, item.data.pressure, item.data.wind, item.data.wspeed, item.data.rain, item.data.snow, item.data.probability, item.data.sky}
		self.putMeasureMessageToArray(arr, from, item_values)
		for i := 0; i < values_size; i++ {
			arr[i+from].Time = item.tstamp
		}
		from += values_size
	}
	self.forecast <- arr
}

func StartWeatherCrawler(index int, location string) (model.MMChannel, model.MMsChannel) {
	crawler := &WeatherCrawler{
		Crawler: Crawler{
			index: index,
		},
		location: location,
	}
	crawler.queue = make(model.MMChannel)
	crawler.forecast = make(model.MMsChannel)
	crawler.ticker = time.NewTicker(15 * time.Minute)
	go func() {
		crawler.poll()
		for _ = range crawler.ticker.C {
			crawler.poll()
		}
	}()
	return crawler.queue, crawler.forecast
}

type WeatherInfo struct {
	Title    string
	Forecast []string
}

type WeatherChan chan *WeatherInfo

type InfoMaker struct {
	channel WeatherChan
	device  int
	sensor  int
	db      *data.DBProvider
}

func formatTemp(val float64) string {
	if float64(int(val)) != val {
		return fmt.Sprintf("%.1fC", val)
	}
	return fmt.Sprintf("%.fC", val)
}

func formatWindDirection(val float64) string {
	step := 22.5
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW", "N"}
	value := step
	idx := 1
	for value < 360 {
		if val >= value && val < value+2*step {
			return directions[idx]
		}
		value += 2 * step
		idx++
	}
	return directions[0]
}

func (self *InfoMaker) oneLatest(measure int) (float64, error) {
	// Get one latest measure
	value, _, err := self.db.LatestMeasure(self.device, wundergroundType, self.sensor, measure)
	return value, err
}

func (self *InfoMaker) oneForecast(measure int, hours int) (float64, error) {
	// Get one latest measure
	msec := time.Now().Add(time.Duration(hours)*time.Hour).Unix() * 1000
	value, _, err := self.db.ClosestForecast(self.device, wundergroundType, self.sensor, measure, msec)
	if err != nil {
		return value, err
	}
	// log.Printf("oneForecast %v %v", hours, time.Local())
	return value, err
}

func (self *InfoMaker) makeLatest() (string, error) {
	temp, err := self.oneLatest(1)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	cond, err := self.oneLatest(0)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	wind, err := self.oneLatest(6)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	wdir, err := self.oneLatest(5)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	hum, err := self.oneLatest(3)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	result := fmt.Sprintf("%s %s %v %v %.f%%", formatTemp(temp), code2Text(int(cond)), formatWindDirection(wdir), wind, hum)
	return result, nil
}

func (self *InfoMaker) makeForecastLine(hours int) (string, error) {
	temp, err := self.oneForecast(1, hours)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	cond, err := self.oneForecast(0, hours)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	wind, err := self.oneForecast(6, hours)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	wdir, err := self.oneForecast(5, hours)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	rain, err := self.oneForecast(9, hours)
	if nil != err {
		log.Printf("Failed to get value: %v", err)
		return "", err
	}
	result := fmt.Sprintf("%dH: %s %s %.f%% %v %v", hours, formatTemp(temp), code2Text(int(cond)), rain, formatWindDirection(wdir), wind)
	return result, nil
}

func (self *InfoMaker) Make() {
	latest, err := self.makeLatest()
	if nil != err {
		log.Printf("Error making latest weather info: %v", err)
		return
	}
	// log.Printf("Weather info message %v", latest)
	forecastLines := make([]string, len(forecastHours))
	for i, hours := range forecastHours {
		line, err := self.makeForecastLine(hours)
		if nil != err {
			log.Printf("Error making forecast line: %v", err)
			return
		}
		// log.Printf("Forecast: %v %v", i, line)
		forecastLines[i] = line
	}
	self.channel <- &WeatherInfo{
		Title:    latest,
		Forecast: forecastLines,
	}
}

func StartWeatherNotifier(db *data.DBProvider, device int, sensor int, minutes int) WeatherChan {
	maker := &InfoMaker{
		device: device,
		sensor: sensor,
		db:     db,
	}
	maker.channel = make(WeatherChan)
	go func() {
		maker.Make()
		for _ = range time.Tick(time.Duration(minutes) * time.Minute) {
			maker.Make()
		}
	}()
	return maker.channel
}
