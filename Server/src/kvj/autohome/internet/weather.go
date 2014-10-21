package internet

import (
	"kvj/autohome/model"
	"log"
	"time"
)

type Crawler struct {
	queue chan *model.MeasureMessage
}

type WeatherCrawler struct {
	Crawler
	Location string
}
