package internet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
