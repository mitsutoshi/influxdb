package influxdb

import (
	"sync"
	"time"
)

type MemAgent struct {
	client   InfluxDBClient
	records  []string
	lock     *sync.Mutex
	interval time.Duration
}

type LogfileAgent struct {
	client   InfluxDBClient
	records  []string
	lock     *sync.Mutex
	interval time.Duration
}
